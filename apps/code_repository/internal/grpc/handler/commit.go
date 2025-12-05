package handler

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

// 提交记录（Commit）CRUD 处理器
func (h *CodeRepositoryGRPCHandler) CreateCommit(ctx context.Context, req *coderepositoryv1.CreateCommitRequest) (*coderepositoryv1.CreateCommitResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.BranchId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and branch_id are required")
	}
	if req.Hash == "" {
		return nil, status.Errorf(codes.InvalidArgument, "hash is required")
	}
	var authoredAt *time.Time
	if req.AuthoredAt != nil {
		t := req.AuthoredAt.AsTime()
		authoredAt = &t
	}
	var committedAt *time.Time
	if req.CommittedAt != nil {
		t := req.CommittedAt.AsTime()
		committedAt = &t
	}
	cm, err := h.session.CreateCommit(ctx, req.ProjectId, req.RepositoryId, req.BranchId, req.Hash, req.Message, req.AuthorName, req.AuthorEmail, authoredAt, req.CommitterName, req.CommitterEmail, committedAt)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.CreateCommitResponse{Commit: cm}, nil
}

func (h *CodeRepositoryGRPCHandler) GetCommit(ctx context.Context, req *coderepositoryv1.GetCommitRequest) (*coderepositoryv1.GetCommitResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.CommitId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and commit_id are required")
	}
	cm, err := h.session.GetCommit(ctx, req.ProjectId, req.RepositoryId, req.CommitId)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.GetCommitResponse{Commit: cm}, nil
}

func (h *CodeRepositoryGRPCHandler) ListCommits(ctx context.Context, req *coderepositoryv1.ListCommitsRequest) (*coderepositoryv1.ListCommitsResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.BranchId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and branch_id are required")
	}
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	items, total, totalPages, err := h.session.ListCommits(ctx, req.ProjectId, req.RepositoryId, req.BranchId, page, pageSize)
	if err != nil {
		return nil, err
	}
	pagination := &coderepositoryv1.ListCommitsResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &coderepositoryv1.ListCommitsResponse{Data: items, Pagination: pagination}, nil
}

func (h *CodeRepositoryGRPCHandler) UpdateCommit(ctx context.Context, req *coderepositoryv1.UpdateCommitRequest) (*coderepositoryv1.UpdateCommitResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.CommitId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and commit_id are required")
	}
	cm, err := h.session.UpdateCommit(ctx, req.ProjectId, req.RepositoryId, req.CommitId, req.Message)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.UpdateCommitResponse{Commit: cm}, nil
}

func (h *CodeRepositoryGRPCHandler) DeleteCommit(ctx context.Context, req *coderepositoryv1.DeleteCommitRequest) (*coderepositoryv1.DeleteCommitResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 || req.CommitId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id, repository_id and commit_id are required")
	}
	if err := h.session.DeleteCommit(ctx, req.ProjectId, req.RepositoryId, req.CommitId); err != nil {
		return nil, err
	}
	return &coderepositoryv1.DeleteCommitResponse{Success: true}, nil
}
