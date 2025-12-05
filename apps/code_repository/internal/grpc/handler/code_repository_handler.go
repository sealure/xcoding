package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xcoding/apps/code_repository/internal/service"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

type CodeRepositoryGRPCHandler struct {
	coderepositoryv1.UnimplementedCodeRepositoryServiceServer
	session service.CodeRepositoryService
}

func NewCodeRepositoryGRPCHandler(svc service.CodeRepositoryService) *CodeRepositoryGRPCHandler {
	return &CodeRepositoryGRPCHandler{session: svc}
}

// 仓库创建：校验必填字段，默认分支缺省为 main
func (h *CodeRepositoryGRPCHandler) CreateRepository(ctx context.Context, req *coderepositoryv1.CreateRepositoryRequest) (*coderepositoryv1.CreateRepositoryResponse, error) {
	if req.ProjectId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id is required")
	}
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.GitUrl == "" {
		return nil, status.Errorf(codes.InvalidArgument, "git_url is required")
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	repo, err := h.session.CreateRepository(ctx, req.ProjectId, req.Name, req.Description, req.GitUrl, req.Branch, req.AuthType, req.GitUsername, req.GitPassword, req.GitSshKey)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.CreateRepositoryResponse{Repository: repo}, nil
}

// 仓库详情查询：需提供项目ID与仓库ID
func (h *CodeRepositoryGRPCHandler) GetRepository(ctx context.Context, req *coderepositoryv1.GetRepositoryRequest) (*coderepositoryv1.GetRepositoryResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	repo, err := h.session.GetRepository(ctx, req.ProjectId, req.RepositoryId)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.GetRepositoryResponse{Repository: repo}, nil
}

// 仓库列表：分页查询并返回分页元数据
func (h *CodeRepositoryGRPCHandler) ListRepositories(ctx context.Context, req *coderepositoryv1.ListRepositoriesRequest) (*coderepositoryv1.ListRepositoriesResponse, error) {
	if req.ProjectId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id is required")
	}
	// 分页参数处理：提供默认值与边界校验
	// page 必须 >= 1；page_size 限制在 [1, maxPageSize]
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	// 调用服务层进行分页查询，并回填 pagination 信息
	items, total, totalPages, err := h.session.ListRepositories(ctx, req.ProjectId, page, pageSize)
	if err != nil {
		return nil, err
	}
	// 构建响应的分页元数据
	pagination := &coderepositoryv1.ListRepositoriesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &coderepositoryv1.ListRepositoriesResponse{Data: items, Pagination: pagination}, nil
}

// 仓库更新：按需更新字段（空值不覆盖）
func (h *CodeRepositoryGRPCHandler) UpdateRepository(ctx context.Context, req *coderepositoryv1.UpdateRepositoryRequest) (*coderepositoryv1.UpdateRepositoryResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.GitUrl != "" {
		updates["git_url"] = req.GitUrl
	}
	if req.Branch != "" {
		updates["branch"] = req.Branch
	}
	if req.AuthType != coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		updates["auth_type"] = req.AuthType
	}
	if req.GitUsername != "" {
		updates["git_username"] = req.GitUsername
	}
	if req.GitPassword != "" {
		updates["git_password"] = req.GitPassword
	}
	if req.GitSshKey != "" {
		updates["git_ssh_key"] = req.GitSshKey
	}
	updates["is_active"] = req.IsActive
	repo, err := h.session.UpdateRepository(ctx, req.ProjectId, req.RepositoryId, updates)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.UpdateRepositoryResponse{Repository: repo}, nil
}

// 仓库删除：软/硬删除由服务层控制
func (h *CodeRepositoryGRPCHandler) DeleteRepository(ctx context.Context, req *coderepositoryv1.DeleteRepositoryRequest) (*coderepositoryv1.DeleteRepositoryResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	if err := h.session.DeleteRepository(ctx, req.ProjectId, req.RepositoryId); err != nil {
		return nil, err
	}
	return &coderepositoryv1.DeleteRepositoryResponse{Success: true}, nil
}

// 仓库连通性测试：校验凭据与网络可达
func (h *CodeRepositoryGRPCHandler) TestRepositoryConnection(ctx context.Context, req *coderepositoryv1.TestRepositoryConnectionRequest) (*coderepositoryv1.TestRepositoryConnectionResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	ok, msg, err := h.session.TestRepositoryConnection(ctx, req.ProjectId, req.RepositoryId)
	if err != nil {
		return nil, err
	}
	return &coderepositoryv1.TestRepositoryConnectionResponse{Success: ok, Message: msg}, nil
}

// 仓库分支列表：分页查询并返回分页元数据
func (h *CodeRepositoryGRPCHandler) GetRepositoryBranches(ctx context.Context, req *coderepositoryv1.GetRepositoryBranchesRequest) (*coderepositoryv1.GetRepositoryBranchesResponse, error) {
	if req.ProjectId == 0 || req.RepositoryId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "project_id and repository_id are required")
	}
	// 分页参数处理：提供默认值与边界校验
	// page 必须 >= 1；page_size 限制在 [1, maxPageSize]
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	// 调用服务层进行分页查询，并回填 pagination 信息
	branches, total, totalPages, err := h.session.GetRepositoryBranches(ctx, req.ProjectId, req.RepositoryId, page, pageSize)
	if err != nil {
		return nil, err
	}
	// 构建响应的分页元数据
	pagination := &coderepositoryv1.GetRepositoryBranchesResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}
	return &coderepositoryv1.GetRepositoryBranchesResponse{Data: branches, Pagination: pagination}, nil
}

// 分支处理器已迁移至 branch.go

// 提交处理器已迁移至 commit.go
