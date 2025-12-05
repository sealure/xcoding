package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

// 分支 CRUD 处理器
func (h *CodeRepositoryGRPCHandler) CreateBranch(ctx context.Context, req *coderepositoryv1.CreateBranchRequest) (*coderepositoryv1.CreateBranchResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	br, err := h.session.CreateBranch(ctx, req.ProjectId, req.RepositoryId, req.Name, req.IsDefault)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.CreateBranchResponse{Branch: br}, nil
}

func (h *CodeRepositoryGRPCHandler) GetBranch(ctx context.Context, req *coderepositoryv1.GetBranchRequest) (*coderepositoryv1.GetBranchResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.BranchId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and branch_id are required")
	}
	br, err := h.session.GetBranch(ctx, req.ProjectId, req.RepositoryId, req.BranchId)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.GetBranchResponse{Branch: br}, nil
}

func (h *CodeRepositoryGRPCHandler) ListBranches(ctx context.Context, req *coderepositoryv1.ListBranchesRequest) (*coderepositoryv1.ListBranchesResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	items, total, totalPages, err := h.session.ListBranches(ctx, req.ProjectId, req.RepositoryId, page, pageSize)
	if err != nil {
		return nil, err
	}
	pagination := &coderepositoryv1.ListBranchesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &coderepositoryv1.ListBranchesResponse{Data: items, Pagination: pagination}, nil
}

func (h *CodeRepositoryGRPCHandler) UpdateBranch(ctx context.Context, req *coderepositoryv1.UpdateBranchRequest) (*coderepositoryv1.UpdateBranchResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.BranchId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and branch_id are required")
	}
	br, err := h.session.UpdateBranch(ctx, req.ProjectId, req.RepositoryId, req.BranchId, req.Name, req.IsDefault)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.UpdateBranchResponse{Branch: br}, nil
}

func (h *CodeRepositoryGRPCHandler) DeleteBranch(ctx context.Context, req *coderepositoryv1.DeleteBranchRequest) (*coderepositoryv1.DeleteBranchResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.BranchId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and branch_id are required")
	}
	if err := h.session.DeleteBranch(ctx, req.ProjectId, req.RepositoryId, req.BranchId); err != nil {
		return nil, err
	}
	return &coderepositoryv1.DeleteBranchResponse{Success: true}, nil
}
