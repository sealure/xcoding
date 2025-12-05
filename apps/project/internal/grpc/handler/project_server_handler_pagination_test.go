package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	projectv1 "xcoding/gen/go/project/v1"
)

// fake service for ProjectService
type fakeProjectService struct{}

func (f *fakeProjectService) CreateProject(ctx context.Context, req *projectv1.CreateProjectRequest, ownerID uint64) (*projectv1.Project, error) {
	return nil, nil
}
func (f *fakeProjectService) GetProject(ctx context.Context, projectID uint64) (*projectv1.Project, error) {
	return nil, nil
}
func (f *fakeProjectService) ListProjects(ctx context.Context, page, pageSize int32, ownerID *uint64, all bool) ([]*projectv1.Project, int32, int32, error) {
	return []*projectv1.Project{}, 0, 0, nil
}
func (f *fakeProjectService) UpdateProject(ctx context.Context, req *projectv1.UpdateProjectRequest, ownerID uint64) (*projectv1.Project, error) {
	return nil, nil
}
func (f *fakeProjectService) DeleteProject(ctx context.Context, projectID uint64) error { return nil }
func (f *fakeProjectService) AddMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error) {
	return nil, nil
}
func (f *fakeProjectService) ListMembers(ctx context.Context, projectID uint64, page, pageSize int32) ([]*projectv1.ProjectMember, int32, int32, error) {
	return []*projectv1.ProjectMember{}, 0, 0, nil
}
func (f *fakeProjectService) UpdateMember(ctx context.Context, projectID, userID uint64, role projectv1.ProjectMemberRole) (*projectv1.ProjectMember, error) {
	return nil, nil
}
func (f *fakeProjectService) RemoveMember(ctx context.Context, projectID, userID uint64) error {
	return nil
}

func withUser(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{"x-user-id": "1"})
	return metadata.NewIncomingContext(ctx, md)
}

func TestListProjects_PaginationValidation(t *testing.T) {
	h := NewProjectGRPCHandler(&fakeProjectService{})
	ctx := withUser(context.Background())

	// invalid page
	_, err := h.ListProjects(ctx, &projectv1.ListProjectsRequest{All: false, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize
	_, err = h.ListProjects(ctx, &projectv1.ListProjectsRequest{All: false, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListProjects(ctx, &projectv1.ListProjectsRequest{All: false, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}

func TestListProjectMembers_PaginationValidation(t *testing.T) {
	h := NewProjectGRPCHandler(&fakeProjectService{})
	ctx := withUser(context.Background())

	// invalid page
	_, err := h.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize
	_, err = h.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}
