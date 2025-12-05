package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	artifactv1 "xcoding/gen/go/artifact/v1"
)

// Namespace operations
func (h *ArtifactGRPCHandler) CreateNamespace(ctx context.Context, req *artifactv1.CreateNamespaceRequest) (*artifactv1.CreateNamespaceResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.RegistryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "registry_id is required")
	}
	namespace, err := h.artifactService.CreateNamespace(ctx, req.Name, req.Description, req.RegistryId)
	if err != nil {
		return nil, err
	}
	return &artifactv1.CreateNamespaceResponse{Namespace: namespace}, nil
}

func (h *ArtifactGRPCHandler) GetNamespace(ctx context.Context, req *artifactv1.GetNamespaceRequest) (*artifactv1.GetNamespaceResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	namespace, err := h.artifactService.GetNamespace(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &artifactv1.GetNamespaceResponse{Namespace: namespace}, nil
}

func (h *ArtifactGRPCHandler) UpdateNamespace(ctx context.Context, req *artifactv1.UpdateNamespaceRequest) (*artifactv1.UpdateNamespaceResponse, error) {
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
	namespace, err := h.artifactService.UpdateNamespace(ctx, req.Id, updates)
	if err != nil {
		return nil, err
	}
	return &artifactv1.UpdateNamespaceResponse{Namespace: namespace}, nil
}

func (h *ArtifactGRPCHandler) ListNamespaces(ctx context.Context, req *artifactv1.ListNamespacesRequest) (*artifactv1.ListNamespacesResponse, error) {
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	namespaces, total, totalPages, err := h.artifactService.ListNamespaces(ctx, page, pageSize, req.RegistryId)
	if err != nil {
		return nil, err
	}
	pagination := &artifactv1.ListNamespacesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &artifactv1.ListNamespacesResponse{Data: namespaces, Pagination: pagination}, nil
}

func (h *ArtifactGRPCHandler) DeleteNamespace(ctx context.Context, req *artifactv1.DeleteNamespaceRequest) (*artifactv1.DeleteNamespaceResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	if err := h.artifactService.DeleteNamespace(ctx, req.Id); err != nil {
		return nil, err
	}
	return &artifactv1.DeleteNamespaceResponse{Success: true}, nil
}
