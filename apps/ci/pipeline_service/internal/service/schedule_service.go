package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/ci/pipeline_service/internal/models"
	civ1 "xcoding/gen/go/ci/v1"
)

// 日程计划 CRUD（创建、查询、更新、删除）
func (s *pipelineService) CreateSchedule(ctx context.Context, req *civ1.CreatePipelineScheduleRequest) (*civ1.CreatePipelineScheduleResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var p models.Pipeline
	if err := s.db.WithContext(ctx).First(&p, req.GetPipelineId()).Error; err != nil {
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
		if err := s.ensureOwnerOrAdmin(ctx, p.ProjectID, actorID); err != nil {
			return nil, err
		}
	}

	sdl := models.PipelineSchedule{
		PipelineID: p.ID,
		Cron:       req.GetCron(),
		Timezone:   req.GetTimezone(),
		Enabled:    req.GetEnabled(),
	}
	if sdl.Timezone == "" {
		sdl.Timezone = "UTC"
	}
	if err := s.db.WithContext(ctx).Select("PipelineID", "Cron", "Timezone", "Enabled").Create(&sdl).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create schedule: %v", err)
	}
	return &civ1.CreatePipelineScheduleResponse{Schedule: sdl.ToProto()}, nil
}

func (s *pipelineService) ListSchedules(ctx context.Context, req *civ1.ListPipelineSchedulesRequest) (*civ1.ListPipelineSchedulesResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var p models.Pipeline
	if err := s.db.WithContext(ctx).First(&p, req.GetPipelineId()).Error; err != nil {
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
		ok, perr := s.isMemberOrHigher(ctx, p.ProjectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to access schedules")
		}
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	q := s.db.WithContext(ctx).Model(&models.PipelineSchedule{}).Where("pipeline_id = ?", p.ID)
	if err := q.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count schedules: %v", err)
	}
	var items []models.PipelineSchedule
	if err := q.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&items).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list schedules: %v", err)
	}
	data := make([]*civ1.PipelineSchedule, len(items))
	for i := range items {
		data[i] = items[i].ToProto()
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return &civ1.ListPipelineSchedulesResponse{
		Data:       data,
		Pagination: &civ1.ListPipelineSchedulesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: int32(total), TotalPages: totalPages},
	}, nil
}

func (s *pipelineService) UpdateSchedule(ctx context.Context, req *civ1.UpdatePipelineScheduleRequest) (*civ1.UpdatePipelineScheduleResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var p models.Pipeline
	if err := s.db.WithContext(ctx).First(&p, req.GetPipelineId()).Error; err != nil {
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
		if err := s.ensureOwnerOrAdmin(ctx, p.ProjectID, actorID); err != nil {
			return nil, err
		}
	}

	var sdl models.PipelineSchedule
	if err := s.db.WithContext(ctx).Where("pipeline_id = ? AND id = ?", p.ID, req.GetScheduleId()).First(&sdl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "schedule not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get schedule: %v", err)
	}
	if v := req.GetCron(); v != "" {
		sdl.Cron = v
	}
	if v := req.GetTimezone(); v != "" {
		sdl.Timezone = v
	}
	sdl.Enabled = req.GetEnabled()
	if err := s.db.WithContext(ctx).Save(&sdl).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update schedule: %v", err)
	}
	return &civ1.UpdatePipelineScheduleResponse{Schedule: sdl.ToProto()}, nil
}

func (s *pipelineService) DeleteSchedule(ctx context.Context, req *civ1.DeletePipelineScheduleRequest) (*civ1.DeletePipelineScheduleResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request nil")
	}
	var p models.Pipeline
	if err := s.db.WithContext(ctx).First(&p, req.GetPipelineId()).Error; err != nil {
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
		if err := s.ensureOwnerOrAdmin(ctx, p.ProjectID, actorID); err != nil {
			return nil, err
		}
	}
	if err := s.db.WithContext(ctx).Where("pipeline_id = ? AND id = ?", p.ID, req.GetScheduleId()).Delete(&models.PipelineSchedule{}).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete schedule: %v", err)
	}
	return &civ1.DeletePipelineScheduleResponse{Success: true}, nil
}
