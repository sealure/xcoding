package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/ci/pipeline_service/internal/models"
	civ1 "xcoding/gen/go/ci/v1"
	projectv1 "xcoding/gen/go/project/v1"
	"xcoding/pkg/auth"
)

// PipelineService 定义 CI 流水线及其相关操作的业务接口
type PipelineService interface {
	CreatePipeline(ctx context.Context, req *civ1.CreatePipelineRequest) (*civ1.CreatePipelineResponse, error)
	GetPipeline(ctx context.Context, req *civ1.GetPipelineRequest) (*civ1.GetPipelineResponse, error)
	ListPipelines(ctx context.Context, req *civ1.ListPipelinesRequest) (*civ1.ListPipelinesResponse, error)
	UpdatePipeline(ctx context.Context, req *civ1.UpdatePipelineRequest) (*civ1.UpdatePipelineResponse, error)
	DeletePipeline(ctx context.Context, req *civ1.DeletePipelineRequest) (*civ1.DeletePipelineResponse, error)

	CreateSchedule(ctx context.Context, req *civ1.CreatePipelineScheduleRequest) (*civ1.CreatePipelineScheduleResponse, error)
	ListSchedules(ctx context.Context, req *civ1.ListPipelineSchedulesRequest) (*civ1.ListPipelineSchedulesResponse, error)
	UpdateSchedule(ctx context.Context, req *civ1.UpdatePipelineScheduleRequest) (*civ1.UpdatePipelineScheduleResponse, error)
	DeleteSchedule(ctx context.Context, req *civ1.DeletePipelineScheduleRequest) (*civ1.DeletePipelineScheduleResponse, error)

	StartPipelineBuild(ctx context.Context, req *civ1.StartPipelineBuildRequest) (*civ1.StartPipelineBuildResponse, error)
}

type pipelineService struct {
	db             *gorm.DB
	projectClient  projectv1.ProjectServiceClient
	executorClient civ1.ExecutorServiceClient
}

func NewPipelineService(db *gorm.DB, projectClient projectv1.ProjectServiceClient, executorClient civ1.ExecutorServiceClient) PipelineService {
	return &pipelineService{db: db, projectClient: projectClient, executorClient: executorClient}
}

// ==== 权限辅助函数 ====
func getUserIDFromCtx(ctx context.Context) (uint64, error)   { return auth.GetUserIDFromCtx(ctx) }
func getUsernameFromCtx(ctx context.Context) (string, error) { return auth.GetUsernameFromCtx(ctx) }
func isUserRoleSuperAdmin(ctx context.Context) bool          { return auth.IsUserRoleSuperAdmin(ctx) }

func (s *pipelineService) isMemberOrHigher(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	p := resp.GetProject()
	if p == nil {
		return false, status.Errorf(codes.NotFound, "project not found")
	}
	if p.OwnerId == actorID {
		return true, nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			return true, nil
		}
	}
	return false, nil
}

func (s *pipelineService) ensureOwnerOrAdmin(ctx context.Context, projectID uint64, actorID uint64) error {
	if isUserRoleSuperAdmin(ctx) {
		return nil
	}
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	p := resp.GetProject()
	if p == nil {
		return status.Errorf(codes.NotFound, "project not found")
	}
	if p.OwnerId == actorID {
		return nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			role := m.GetRole()
			if role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER || role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN {
				return nil
			}
		}
	}
	return status.Errorf(codes.PermissionDenied, "only owner or admin can perform this action")
}

// ==== 流水线 CRUD（创建、查询、更新、删除） ====
func (s *pipelineService) CreatePipeline(ctx context.Context, req *civ1.CreatePipelineRequest) (*civ1.CreatePipelineResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	projectID := req.GetProjectId()
	if projectID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id is required")
	}
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return nil, err
		}
	}

	// 唯一性：同一项目下的名称需唯一
	var existing models.Pipeline
	if err := s.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, req.GetName()).First(&existing).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.Internal, "failed to check pipeline: %v", err)
		}
	} else {
		return nil, status.Errorf(codes.AlreadyExists, "pipeline already exists in project")
	}

	m := models.Pipeline{
		ProjectID:    projectID,
		Name:         req.GetName(),
		Description:  req.GetDescription(),
		WorkflowYAML: req.GetWorkflowYaml(),
		IsActive:     req.GetIsActive(),
	}
	// 显式选择字段以确保布尔值 false 被正确持久化
	if err := s.db.WithContext(ctx).Select("ProjectID", "Name", "Description", "WorkflowYAML", "IsActive").Create(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create pipeline: %v", err)
	}
	return &civ1.CreatePipelineResponse{Pipeline: m.ToProto()}, nil
}

func (s *pipelineService) GetPipeline(ctx context.Context, req *civ1.GetPipelineRequest) (*civ1.GetPipelineResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var m models.Pipeline
	if err := s.db.WithContext(ctx).First(&m, req.GetPipelineId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "pipeline not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get pipeline: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		actorID, err := getUserIDFromCtx(ctx)
		if err != nil {
			return nil, err
		}
		ok, perr := s.isMemberOrHigher(ctx, m.ProjectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to access pipeline")
		}
	}
	return &civ1.GetPipelineResponse{Pipeline: m.ToProto()}, nil
}

func (s *pipelineService) ListPipelines(ctx context.Context, req *civ1.ListPipelinesRequest) (*civ1.ListPipelinesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	page := int32(1)
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	q := s.db.WithContext(ctx).Model(&models.Pipeline{})
	if pid := req.GetProjectId(); pid > 0 {
		q = q.Where("project_id = ?", pid)
	}
	if name := req.GetName(); name != "" {
		q = q.Where("name ILIKE ?", "%"+name+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count pipelines: %v", err)
	}

	var items []models.Pipeline
	if err := q.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&items).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list pipelines: %v", err)
	}
	data := make([]*civ1.Pipeline, len(items))
	for i := range items {
		data[i] = items[i].ToProto()
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return &civ1.ListPipelinesResponse{
		Data:       data,
		Pagination: &civ1.ListPipelinesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: int32(total), TotalPages: totalPages},
	}, nil
}

func (s *pipelineService) UpdatePipeline(ctx context.Context, req *civ1.UpdatePipelineRequest) (*civ1.UpdatePipelineResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var m models.Pipeline
	if err := s.db.WithContext(ctx).First(&m, req.GetPipelineId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "pipeline not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get pipeline: %v", err)
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, m.ProjectID, actorID); err != nil {
			return nil, err
		}
	}

	if v := req.GetName(); v != "" {
		m.Name = v
	}
	if v := req.GetDescription(); v != "" {
		m.Description = v
	}
	if v := req.GetProjectId(); v != 0 {
		m.ProjectID = v
	}
	if v := req.GetWorkflowYaml(); v != "" {
		m.WorkflowYAML = v
	}
	// 保留显式的 false 值
	m.IsActive = req.GetIsActive()

	if err := s.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update pipeline: %v", err)
	}
	return &civ1.UpdatePipelineResponse{Pipeline: m.ToProto()}, nil
}

func (s *pipelineService) DeletePipeline(ctx context.Context, req *civ1.DeletePipelineRequest) (*civ1.DeletePipelineResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var m models.Pipeline
	if err := s.db.WithContext(ctx).First(&m, req.GetPipelineId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "pipeline not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get pipeline: %v", err)
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, m.ProjectID, actorID); err != nil {
			return nil, err
		}
	}
	if err := s.db.WithContext(ctx).Delete(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete pipeline: %v", err)
	}
	return &civ1.DeletePipelineResponse{Success: true}, nil
}

// ==== 日程计划 CRUD ====
// 日程与构建的具体实现已迁移至 schedule_service.go 和 build_service.go
