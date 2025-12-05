package handler

import (
	"context"

	"xcoding/apps/ci/pipeline_service/internal/service"
	civ1 "xcoding/gen/go/ci/v1"
)

type PipelineGRPCHandler struct {
	civ1.UnimplementedPipelineServiceServer
	pipelineService service.PipelineService
}

func NewPipelineGRPCHandler(svc service.PipelineService) *PipelineGRPCHandler {
	return &PipelineGRPCHandler{pipelineService: svc}
}

// 流水线 CRUD 的 gRPC 接口
func (h *PipelineGRPCHandler) CreatePipeline(ctx context.Context, req *civ1.CreatePipelineRequest) (*civ1.CreatePipelineResponse, error) {
	return h.pipelineService.CreatePipeline(ctx, req)
}
func (h *PipelineGRPCHandler) GetPipeline(ctx context.Context, req *civ1.GetPipelineRequest) (*civ1.GetPipelineResponse, error) {
	return h.pipelineService.GetPipeline(ctx, req)
}
func (h *PipelineGRPCHandler) ListPipelines(ctx context.Context, req *civ1.ListPipelinesRequest) (*civ1.ListPipelinesResponse, error) {
	return h.pipelineService.ListPipelines(ctx, req)
}
func (h *PipelineGRPCHandler) UpdatePipeline(ctx context.Context, req *civ1.UpdatePipelineRequest) (*civ1.UpdatePipelineResponse, error) {
	return h.pipelineService.UpdatePipeline(ctx, req)
}
func (h *PipelineGRPCHandler) DeletePipeline(ctx context.Context, req *civ1.DeletePipelineRequest) (*civ1.DeletePipelineResponse, error) {
	return h.pipelineService.DeletePipeline(ctx, req)
}
