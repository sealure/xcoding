package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 制品仓库（Registry）相关操作
func (h *ArtifactGRPCHandler) CreateRegistry(ctx context.Context, req *artifactv1.CreateRegistryRequest) (*artifactv1.CreateRegistryResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.Url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "url is required")
	}
	if req.ProjectId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id is required")
	}
	registry, err := h.artifactService.CreateRegistry(ctx, req.Name, req.Url, req.Description, req.Username, req.Password, req.IsPublic, req.ProjectId, req.ArtifactType, req.ArtifactSource)
	if err != nil {
		return nil, err
	}
	return &artifactv1.CreateRegistryResponse{Registry: registry}, nil
}

func (h *ArtifactGRPCHandler) GetRegistry(ctx context.Context, req *artifactv1.GetRegistryRequest) (*artifactv1.GetRegistryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	registry, err := h.artifactService.GetRegistry(ctx, req.Id, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &artifactv1.GetRegistryResponse{Registry: registry}, nil
}

func (h *ArtifactGRPCHandler) UpdateRegistry(ctx context.Context, req *artifactv1.UpdateRegistryRequest) (*artifactv1.UpdateRegistryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Url != "" {
		updates["url"] = req.Url
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["is_public"] = req.IsPublic
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	updates["artifact_type"] = int32(req.ArtifactType)
	updates["artifact_source"] = int32(req.ArtifactSource)
	registry, err := h.artifactService.UpdateRegistry(ctx, req.Id, updates, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &artifactv1.UpdateRegistryResponse{Registry: registry}, nil
}

func (h *ArtifactGRPCHandler) ListRegistries(ctx context.Context, req *artifactv1.ListRegistriesRequest) (*artifactv1.ListRegistriesResponse, error) {
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	registries, total, totalPages, err := h.artifactService.ListRegistries(ctx, page, pageSize, req.ProjectId)
	if err != nil {
		return nil, err
	}
	pagination := &artifactv1.ListRegistriesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &artifactv1.ListRegistriesResponse{Data: registries, Pagination: pagination}, nil
}

func (h *ArtifactGRPCHandler) DeleteRegistry(ctx context.Context, req *artifactv1.DeleteRegistryRequest) (*artifactv1.DeleteRegistryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	if err := h.artifactService.DeleteRegistry(ctx, req.Id, req.ProjectId); err != nil {
		return nil, err
	}
	return &artifactv1.DeleteRegistryResponse{Success: true}, nil
}
