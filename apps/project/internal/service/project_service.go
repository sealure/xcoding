package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	// "strings" // removed, now handled by roles helper

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/project/internal/models"
	projectv1 "xcoding/gen/go/project/v1"
)

type ProjectService interface {
	CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest, ownerID uint64) (*projectv1.Project, error)
	GetProject(ctx context.Context, projectID uint64) (*projectv1.Project, error)
	ListProjects(ctx context.Context, page, pageSize int32, ownerID *uint64, all bool) ([]*projectv1.Project, int32, int32, error)
	UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest, ownerID uint64) (*projectv1.Project, error)
	DeleteProject(ctx context.Context, projectID uint64) error
	AddMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error)
	ListMembers(ctx context.Context, projectID uint64, page, pageSize int32) ([]*projectv1.ProjectMember, int32, int32, error)
	UpdateMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID uint64) error
}

type projectService struct{ db *gorm.DB }

func NewProjectService(db *gorm.DB) ProjectService { return &projectService{db: db} }

func (s *projectService) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest, ownerID uint64) (*projectv1.Project, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	p := models.Project{
		Name:        req.Name,
		Description: req.Description,
		Language:    req.Language,
		Framework:   req.Framework,
		IsPublic:    req.IsPublic,
		Status:      "active",
		OwnerID:     ownerID,
	}
	if err := s.db.WithContext(ctx).Create(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}
	log.Printf("audit: action=create_project project_id=%d owner_id=%d name=%s", p.ID, ownerID, p.Name)
	return p.ToProto(), nil
}

func (s *projectService) GetProject(ctx context.Context, projectID uint64) (*projectv1.Project, error) {
	var p models.Project
	if err := s.db.WithContext(ctx).First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	return p.ToProto(), nil
}

func (s *projectService) ListProjects(ctx context.Context, page, pageSize int32, ownerID *uint64, all bool) ([]*projectv1.Project, int32, int32, error) {
	offset := (page - 1) * pageSize
	var projects []models.Project
	q := s.db.WithContext(ctx).Model(&models.Project{})
	// 为了让数据查询和总数统计共享同一过滤条件，构造一个可复用的查询链
	dataQuery := q

	// 基于 x-user-role 元数据限制 all=true 仅允许管理员使用
	if all {
		var actor uint64
		if id, err := getUserIDFromCtx(ctx); err == nil {
			actor = id
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			vals := md.Get("x-user-role")
			role := ""
			if len(vals) > 0 {
				role = vals[0]
			}
			// 使用辅助函数检查是否为全局管理员
			allowed := isGlobalAdmin(ctx)
			// 回退策略：允许项目级 OWNER 或 ADMIN 成员
			if !allowed && actor != 0 {
				var cnt int64
				if err := s.db.WithContext(ctx).
					Model(&models.ProjectMember{}).
					Where("user_id = ? AND role IN (?, ?)", actor,
						int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN),
						int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER),
					).
					Count(&cnt).Error; err != nil {
					return nil, 0, 0, status.Errorf(codes.Internal, "permission check failed: %v", err)
				}
				if cnt > 0 {
					allowed = true
				}
			}

			if allowed {
				log.Printf("audit: action=list_projects_all result=allowed actor_id=%d role=%s", actor, role)
			} else {
				log.Printf("audit: action=list_projects_all result=denied actor_id=%d role=%s", actor, role)
				return nil, 0, 0, status.Errorf(codes.PermissionDenied, "only owner or admin can list all projects")
			}
		} else {
			log.Printf("audit: action=list_projects_all result=unauthenticated")
			return nil, 0, 0, status.Errorf(codes.Unauthenticated, "missing metadata")
		}
	}

	if !all && ownerID != nil {
		sub := s.db.WithContext(ctx).Model(&models.ProjectMember{}).Select("project_id").Where("user_id = ?", *ownerID)
		dataQuery = dataQuery.Where("owner_id = ? OR id IN (?)", *ownerID, sub)
		if err := dataQuery.Offset(int(offset)).Limit(int(pageSize)).Find(&projects).Error; err != nil {
			return nil, 0, 0, status.Errorf(codes.Internal, "failed to list projects: %v", err)
		}
	} else if !all {
		// 若无法识别用户（ownerID=nil），仍按原逻辑返回全部列表（受上游网关认证约束），但统计与数据保持一致
		if err := dataQuery.Offset(int(offset)).Limit(int(pageSize)).Find(&projects).Error; err != nil {
			return nil, 0, 0, status.Errorf(codes.Internal, "failed to list projects: %v", err)
		}
	} else {
		// all=true 且已授权（在前置校验已确认），返回全部
		if err := dataQuery.Offset(int(offset)).Limit(int(pageSize)).Find(&projects).Error; err != nil {
			return nil, 0, 0, status.Errorf(codes.Internal, "failed to list projects: %v", err)
		}
	}

	var total int64
	// 使用与数据查询一致的过滤条件进行计数，避免权限下返回总页数偏大
	if err := dataQuery.Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count projects: %v", err)
	}

	totalPages := (int32(total) + pageSize - 1) / pageSize

	res := make([]*projectv1.Project, 0, len(projects))
	for _, p := range projects {
		res = append(res, p.ToProto())
	}

	return res, int32(total), totalPages, nil
}

func (s *projectService) UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest, ownerID uint64) (*projectv1.Project, error) {
	var p models.Project
	if err := s.db.WithContext(ctx).First(&p, req.ProjectId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}

	// Allow owner or admin to update project
	ok, err := s.isOwnerOrAdmin(ctx, req.ProjectId, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "permission check failed: %v", err)
	}
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "only owner or admin can update project")
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Language != "" {
		p.Language = req.Language
	}
	if req.Framework != "" {
		p.Framework = req.Framework
	}
	p.IsPublic = req.IsPublic
	if req.Status != "" {
		p.Status = req.Status
	}

	if err := s.db.WithContext(ctx).Save(&p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update project: %v", err)
	}
	log.Printf("audit: action=update_project result=ok project_id=%d actor_id=%d", p.ID, ownerID)
	return p.ToProto(), nil
}

func (s *projectService) DeleteProject(ctx context.Context, projectID uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	ok, err := s.isOwnerOrAdmin(ctx, projectID, actorID)
	if err != nil {
		return status.Errorf(codes.Internal, "permission check failed: %v", err)
	}
	if !ok {
		log.Printf("audit: action=delete_project result=denied project_id=%d actor_id=%d", projectID, actorID)
		return status.Errorf(codes.PermissionDenied, "only owner or admin can delete project")
	}
	if err := s.db.WithContext(ctx).Delete(&models.Project{}, projectID).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete project: %v", err)
	}
	log.Printf("audit: action=delete_project result=ok project_id=%d actor_id=%d", projectID, actorID)
	return nil
}

func (s *projectService) AddMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error) {
	var p models.Project
	if err := s.db.WithContext(ctx).First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	ok, err := s.isOwnerOrAdmin(ctx, projectID, actorID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "permission check failed: %v", err)
	}
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "only owner or admin can manage members")
	}
	var username string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-username"); len(vals) > 0 {
			username = vals[0]
		}
	}
	m := models.ProjectMember{UserID: userID, ProjectID: projectID, Role: int32(role), Username: username}
	if err := s.db.WithContext(ctx).
		Where(&models.ProjectMember{UserID: userID, ProjectID: projectID}).
		Assign(models.ProjectMember{Role: int32(role), Username: username}).
		FirstOrCreate(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add member: %v", err)
	}
	log.Printf("audit: action=add_member project_id=%d actor_id=%d target_user_id=%d role=%d", projectID, actorID, userID, int32(role))
	return m.ToProto(), nil
}

func (s *projectService) UpdateMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	ok, err := s.isOwnerOrAdmin(ctx, projectID, actorID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "permission check failed: %v", err)
	}
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "only owner or admin can manage members")
	}
	var m models.ProjectMember
	if err := s.db.WithContext(ctx).First(&m, &models.ProjectMember{ProjectID: projectID, UserID: userID}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "member not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get member: %v", err)
	}
	m.Role = int32(role)
	if err := s.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update member: %v", err)
	}
	log.Printf("audit: action=update_member project_id=%d actor_id=%d target_user_id=%d role=%d", projectID, actorID, userID, int32(role))
	return m.ToProto(), nil
}

func (s *projectService) RemoveMember(ctx context.Context, projectID, userID uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	ok, err := s.isOwnerOrAdmin(ctx, projectID, actorID)
	if err != nil {
		return status.Errorf(codes.Internal, "permission check failed: %v", err)
	}
	if !ok {
		return status.Errorf(codes.PermissionDenied, "only owner or admin can manage members")
	}
	if err := s.db.WithContext(ctx).Delete(&models.ProjectMember{}, &models.ProjectMember{ProjectID: projectID, UserID: userID}).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to remove member: %v", err)
	}
	log.Printf("audit: action=remove_member project_id=%d actor_id=%d target_user_id=%d", projectID, actorID, userID)
	return nil
}

func (s *projectService) ListMembers(ctx context.Context, projectID uint64, page, pageSize int32) ([]*projectv1.ProjectMember, int32, int32, error) {
	offset := (page - 1) * pageSize
	var members []models.ProjectMember

	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Offset(int(offset)).Limit(int(pageSize)).Find(&members).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list members: %v", err)
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.ProjectMember{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count members: %v", err)
	}

	totalPages := (int32(total) + pageSize - 1) / pageSize
	res := make([]*projectv1.ProjectMember, 0, len(members))
	for _, m := range members {
		res = append(res, m.ToProto())
	}
	return res, int32(total), totalPages, nil
}

// 从上下文的元数据头中提取 ownerID（通常来自 forward-auth 注入）
func GetOwnerIDFromContext(ctx context.Context) (uint64, error) {
	// In real code, we'd parse context with user info injected by interceptors
	return 0, fmt.Errorf("not implemented")
}

// == 新增的辅助方法 ==
func getUserIDFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing x-user-id header")
	}
	id, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid x-user-id: %v", err)
	}
	return id, nil
}

func (s *projectService) isOwnerOrAdmin(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	// Global admins are always authorized
	if isGlobalAdmin(ctx) {
		return true, nil
	}
	var p models.Project
	if err := s.db.WithContext(ctx).First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, status.Errorf(codes.NotFound, "project not found")
		}
		return false, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	if p.OwnerID == actorID {
		return true, nil
	}
	var mem models.ProjectMember
	err := s.db.WithContext(ctx).First(&mem, &models.ProjectMember{ProjectID: projectID, UserID: actorID}).Error
	if err == nil {
		if mem.Role == int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN) || mem.Role == int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER) {
			return true, nil
		}
		return false, nil
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, status.Errorf(codes.Internal, "failed to check member: %v", err)
}
