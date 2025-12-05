package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 仓库相关操作
func (h *ArtifactGRPCHandler) CreateRepository(ctx context.Context, req *artifactv1.CreateRepositoryRequest) (*artifactv1.CreateRepositoryResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.NamespaceId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "namespace_id is required")
	}
	repository, err := h.artifactService.CreateRepository(ctx, req.Name, req.Description, req.NamespaceId, req.IsPublic, req.Path)
	if err != nil {
		return nil, err
	}
	return &artifactv1.CreateRepositoryResponse{Repository: repository}, nil
}

func (h *ArtifactGRPCHandler) GetRepository(ctx context.Context, req *artifactv1.GetRepositoryRequest) (*artifactv1.GetRepositoryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	repository, err := h.artifactService.GetRepository(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &artifactv1.GetRepositoryResponse{Repository: repository}, nil
}

func (h *ArtifactGRPCHandler) UpdateRepository(ctx context.Context, req *artifactv1.UpdateRepositoryRequest) (*artifactv1.UpdateRepositoryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// 支持更新 is_public
	if req.IsPublic {
		updates["is_public"] = req.IsPublic
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	repository, err := h.artifactService.UpdateRepository(ctx, req.Id, updates)
	if err != nil {
		return nil, err
	}
	return &artifactv1.UpdateRepositoryResponse{Repository: repository}, nil
}

func (h *ArtifactGRPCHandler) ListRepositories(ctx context.Context, req *artifactv1.ListRepositoriesRequest) (*artifactv1.ListRepositoriesResponse, error) {
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	repositories, total, totalPages, err := h.artifactService.ListRepositories(ctx, page, pageSize, req.NamespaceId)
	if err != nil {
		return nil, err
	}
	pagination := &artifactv1.ListRepositoriesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &artifactv1.ListRepositoriesResponse{Data: repositories, Pagination: pagination}, nil
}

func (h *ArtifactGRPCHandler) DeleteRepository(ctx context.Context, req *artifactv1.DeleteRepositoryRequest) (*artifactv1.DeleteRepositoryResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	if err := h.artifactService.DeleteRepository(ctx, req.Id); err != nil {
		return nil, err
	}
	return &artifactv1.DeleteRepositoryResponse{Success: true}, nil
}
