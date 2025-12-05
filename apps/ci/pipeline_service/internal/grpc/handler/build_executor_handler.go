package handler

import (
	"context"

	"xcoding/apps/ci/pipeline_service/internal/service"
)

// BuildExecutorHandler wires queue and executor with service
type BuildExecutorHandler struct {
	q service.BuildQueue
}

func NewBuildExecutorHandler(q service.BuildQueue) *BuildExecutorHandler {
	return &BuildExecutorHandler{q: q}
}

// Init registers the queue into service package for StartPipelineBuild usage
func (h *BuildExecutorHandler) Init(ctx context.Context) error {
	service.SetBuildQueue(h.q)
	return nil
}
