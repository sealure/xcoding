package handler

import (
	"xcoding/apps/artifact/internal/service"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

type ArtifactGRPCHandler struct {
	artifactv1.UnimplementedArtifactServiceServer
	artifactService service.ArtifactService
}

func NewArtifactGRPCHandler(artifactService service.ArtifactService) *ArtifactGRPCHandler {
	return &ArtifactGRPCHandler{
		artifactService: artifactService,
	}
}
