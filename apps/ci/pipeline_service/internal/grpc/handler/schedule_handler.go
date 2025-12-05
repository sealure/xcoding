package handler

import (
	"context"

	civ1 "xcoding/gen/go/ci/v1"
)

// 日程计划相关 gRPC 接口
func (h *PipelineGRPCHandler) CreateSchedule(ctx context.Context, req *civ1.CreatePipelineScheduleRequest) (*civ1.CreatePipelineScheduleResponse, error) {
	return h.pipelineService.CreateSchedule(ctx, req)
}

func (h *PipelineGRPCHandler) ListSchedules(ctx context.Context, req *civ1.ListPipelineSchedulesRequest) (*civ1.ListPipelineSchedulesResponse, error) {
	return h.pipelineService.ListSchedules(ctx, req)
}

func (h *PipelineGRPCHandler) UpdateSchedule(ctx context.Context, req *civ1.UpdatePipelineScheduleRequest) (*civ1.UpdatePipelineScheduleResponse, error) {
	return h.pipelineService.UpdateSchedule(ctx, req)
}

func (h *PipelineGRPCHandler) DeleteSchedule(ctx context.Context, req *civ1.DeletePipelineScheduleRequest) (*civ1.DeletePipelineScheduleResponse, error) {
	return h.pipelineService.DeleteSchedule(ctx, req)
}
