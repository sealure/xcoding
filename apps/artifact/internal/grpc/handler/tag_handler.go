package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 标签相关操作
func (h *ArtifactGRPCHandler) CreateTag(ctx context.Context, req *artifactv1.CreateTagRequest) (*artifactv1.CreateTagResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.Digest == "" {
		return nil, status.Errorf(codes.InvalidArgument, "digest is required")
	}
	if req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "repository_id is required")
	}
	tag, err := h.artifactService.CreateTag(ctx, req.Name, req.Digest, int64(req.SizeBytes), req.RepositoryId, false)
	if err != nil {
		return nil, err
	}
	return &artifactv1.CreateTagResponse{Tag: tag}, nil
}

func (h *ArtifactGRPCHandler) GetTag(ctx context.Context, req *artifactv1.GetTagRequest) (*artifactv1.GetTagResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	tag, err := h.artifactService.GetTag(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &artifactv1.GetTagResponse{Tag: tag}, nil
}

func (h *ArtifactGRPCHandler) UpdateTag(ctx context.Context, req *artifactv1.UpdateTagRequest) (*artifactv1.UpdateTagResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Digest != "" {
		updates["digest"] = req.Digest
	}
	if req.SizeBytes > 0 {
		updates["size"] = int64(req.SizeBytes)
	}
	tag, err := h.artifactService.UpdateTag(ctx, req.Id, updates)
	if err != nil {
		return nil, err
	}
	return &artifactv1.UpdateTagResponse{Tag: tag}, nil
}

func (h *ArtifactGRPCHandler) ListTags(ctx context.Context, req *artifactv1.ListTagsRequest) (*artifactv1.ListTagsResponse, error) {
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	tags, total, totalPages, err := h.artifactService.ListTags(ctx, page, pageSize, req.RepositoryId)
	if err != nil {
		return nil, err
	}
	pagination := &artifactv1.ListTagsResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &artifactv1.ListTagsResponse{Data: tags, Pagination: pagination}, nil
}

func (h *ArtifactGRPCHandler) DeleteTag(ctx context.Context, req *artifactv1.DeleteTagRequest) (*artifactv1.DeleteTagResponse, error) {
	if req.Id == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required")
	}
	if err := h.artifactService.DeleteTag(ctx, req.Id); err != nil {
		return nil, err
	}
	return &artifactv1.DeleteTagResponse{Success: true}, nil
}
