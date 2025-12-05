package handler

import (
	"context"

	civ1 "xcoding/gen/go/ci/v1"
)

// 构建相关 gRPC 接口
func (h *PipelineGRPCHandler) StartPipelineBuild(ctx context.Context, req *civ1.StartPipelineBuildRequest) (*civ1.StartPipelineBuildResponse, error) {
	return h.pipelineService.StartPipelineBuild(ctx, req)
}
