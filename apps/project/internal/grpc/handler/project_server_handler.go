package server

import (
	"context"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"xcoding/apps/project/internal/service"
	projectv1 "xcoding/gen/go/project/v1"
)

type ProjectGRPCHandler struct {
	projectv1.UnimplementedProjectServiceServer
	projectService service.ProjectService
}

func NewProjectGRPCHandler(projectService service.ProjectService) *ProjectGRPCHandler {
	return &ProjectGRPCHandler{projectService: projectService}
}

func (h *ProjectGRPCHandler) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest) (*projectv1.CreateProjectResponse, error) {
	ownerID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	// propagate user metadata to response headers for gateway to echo
	propagateUserHeaders(ctx)
	p, err := h.projectService.CreateProject(ctx, req, ownerID)
	if err != nil {
		return nil, err
	}
	return &projectv1.CreateProjectResponse{Project: p}, nil
}

func (h *ProjectGRPCHandler) GetProject(ctx context.Context, req *projectv1.GetProjectRequest) (*projectv1.GetProjectResponse, error) {
	propagateUserHeaders(ctx)
	p, err := h.projectService.GetProject(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &projectv1.GetProjectResponse{Project: p}, nil
}

func (h *ProjectGRPCHandler) ListProjects(ctx context.Context, req *projectv1.ListProjectsRequest) (*projectv1.ListProjectsResponse, error) {
	var ownerID *uint64
	if id, err := getUserIDFromMetadata(ctx); err == nil {
		ownerID = &id
	}
	propagateUserHeaders(ctx)
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	projects, total, totalPages, err := h.projectService.ListProjects(ctx, page, pageSize, ownerID, req.All)
	if err != nil {
		return nil, err
	}
	return &projectv1.ListProjectsResponse{
		Data:       projects,
		Pagination: &projectv1.ListProjectsResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages},
	}, nil
}

func (h *ProjectGRPCHandler) UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest) (*projectv1.UpdateProjectResponse, error) {
	ownerID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	propagateUserHeaders(ctx)
	p, err := h.projectService.UpdateProject(ctx, req, ownerID)
	if err != nil {
		return nil, err
	}
	return &projectv1.UpdateProjectResponse{Project: p}, nil
}

func (h *ProjectGRPCHandler) DeleteProject(ctx context.Context, req *projectv1.DeleteProjectRequest) (*projectv1.DeleteProjectResponse, error) {
	propagateUserHeaders(ctx)
	if err := h.projectService.DeleteProject(ctx, req.ProjectId); err != nil {
		return nil, err
	}
	return &projectv1.DeleteProjectResponse{Success: true}, nil
}

func (h *ProjectGRPCHandler) AddProjectMember(ctx context.Context, req *projectv1.AddProjectMemberRequest) (*projectv1.AddProjectMemberResponse, error) {
	propagateUserHeaders(ctx)
	m, err := h.projectService.AddMember(ctx, req.ProjectId, req.UserId, req.Role)
	if err != nil {
		return nil, err
	}
	return &projectv1.AddProjectMemberResponse{Member: m}, nil
}

func (h *ProjectGRPCHandler) ListProjectMembers(ctx context.Context, req *projectv1.ListProjectMembersRequest) (*projectv1.ListProjectMembersResponse, error) {
	propagateUserHeaders(ctx)
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	members, total, totalPages, err := h.projectService.ListMembers(ctx, req.ProjectId, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &projectv1.ListProjectMembersResponse{
		Data:       members,
		Pagination: &projectv1.ListProjectMembersResponse_Pagination{Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages},
	}, nil
}

func (h *ProjectGRPCHandler) UpdateProjectMember(ctx context.Context, req *projectv1.UpdateProjectMemberRequest) (*projectv1.UpdateProjectMemberResponse, error) {
	propagateUserHeaders(ctx)
	m, err := h.projectService.UpdateMember(ctx, req.ProjectId, req.UserId, req.Role)
	if err != nil {
		return nil, err
	}
	return &projectv1.UpdateProjectMemberResponse{Member: m}, nil
}

func (h *ProjectGRPCHandler) RemoveProjectMember(ctx context.Context, req *projectv1.RemoveProjectMemberRequest) (*projectv1.RemoveProjectMemberResponse, error) {
	propagateUserHeaders(ctx)
	if err := h.projectService.RemoveMember(ctx, req.ProjectId, req.UserId); err != nil {
		return nil, err
	}
	return &projectv1.RemoveProjectMemberResponse{Success: true}, nil
}

func (h *ProjectGRPCHandler) CreateProjectWithUser(ctx context.Context, req *projectv1.CreateProjectWithUserRequest) (*projectv1.CreateProjectWithUserResponse, error) {
	// 简化处理：直接创建项目，忽略用户创建逻辑（可后续集成）
	ownerID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	propagateUserHeaders(ctx)
	p, err := h.projectService.CreateProject(ctx, &projectv1.CreateProjectRequest{
		Name:        req.GetProject().GetName(),
		Description: req.GetProject().GetDescription(),
		Language:    req.GetProject().GetLanguage(),
		Framework:   req.GetProject().GetFramework(),
		IsPublic:    req.GetProject().GetIsPublic(),
	}, ownerID)
	if err != nil {
		return nil, err
	}
	return &projectv1.CreateProjectWithUserResponse{Project: p, User: req.User}, nil
}

func (h *ProjectGRPCHandler) SyncUserPermissions(ctx context.Context, req *projectv1.SyncUserPermissionsRequest) (*projectv1.SyncUserPermissionsResponse, error) {
	propagateUserHeaders(ctx)
	members, _, _, err := h.projectService.ListMembers(ctx, req.ProjectId, 1, 100)
	if err != nil {
		return nil, err
	}
	return &projectv1.SyncUserPermissionsResponse{Members: members}, nil
}

func getUserIDFromMetadata(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing x-user-id header")
	}
	id, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid x-user-id: %v", err)
	}
	return id, nil
}

func propagateUserHeaders(ctx context.Context) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		headerMD := metadata.New(nil)
		for _, k := range []string{"x-user-id", "x-username", "x-user-role", "x-scopes"} {
			vals := md.Get(k)
			if len(vals) > 0 {
				headerMD.Set(k, vals[0])
			}
		}
		// Set header metadata so gateway can echo them via ServerMetadataFromContext
		_ = grpc.SetHeader(ctx, headerMD)
	}
}
