package handler

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

// fake service implementing CodeRepositoryService for tests
type fakeService struct{}

func (f *fakeService) CreateRepository(ctx context.Context, projectID uint64, name, description, gitURL, branch string, authType coderepositoryv1.RepositoryAuthType, gitUsername, gitPassword, gitSSHKey string) (*coderepositoryv1.Repository, error) {
	return nil, nil
}
func (f *fakeService) GetRepository(ctx context.Context, projectID, repositoryID uint64) (*coderepositoryv1.Repository, error) {
	return nil, nil
}
func (f *fakeService) ListRepositories(ctx context.Context, projectID uint64, page, pageSize int32) ([]*coderepositoryv1.Repository, int32, int32, error) {
	return []*coderepositoryv1.Repository{}, 0, 0, nil
}
func (f *fakeService) UpdateRepository(ctx context.Context, projectID, repositoryID uint64, updates map[string]interface{}) (*coderepositoryv1.Repository, error) {
	return nil, nil
}
func (f *fakeService) DeleteRepository(ctx context.Context, projectID, repositoryID uint64) error {
	return nil
}
func (f *fakeService) TestRepositoryConnection(ctx context.Context, projectID, repositoryID uint64) (bool, string, error) {
	return true, "", nil
}
func (f *fakeService) GetRepositoryBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]string, int32, int32, error) {
	return []string{}, 0, 0, nil
}

// Branch
func (f *fakeService) CreateBranch(ctx context.Context, projectID, repositoryID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error) {
	return nil, nil
}
func (f *fakeService) GetBranch(ctx context.Context, projectID, repositoryID, branchID uint64) (*coderepositoryv1.Branch, error) {
	return nil, nil
}
func (f *fakeService) ListBranches(ctx context.Context, projectID, repositoryID uint64, page, pageSize int32) ([]*coderepositoryv1.Branch, int32, int32, error) {
	return []*coderepositoryv1.Branch{}, 0, 0, nil
}
func (f *fakeService) UpdateBranch(ctx context.Context, projectID, repositoryID, branchID uint64, name string, isDefault bool) (*coderepositoryv1.Branch, error) {
	return nil, nil
}
func (f *fakeService) DeleteBranch(ctx context.Context, projectID, repositoryID, branchID uint64) error {
	return nil
}

// Commit
func (f *fakeService) CreateCommit(ctx context.Context, projectID, repositoryID, branchID uint64, hash, message, authorName, authorEmail string, authoredAt *time.Time, committerName, committerEmail string, committedAt *time.Time) (*coderepositoryv1.Commit, error) {
	return nil, nil
}
func (f *fakeService) GetCommit(ctx context.Context, projectID, repositoryID, commitID uint64) (*coderepositoryv1.Commit, error) {
	return nil, nil
}
func (f *fakeService) ListCommits(ctx context.Context, projectID, repositoryID, branchID uint64, page, pageSize int32) ([]*coderepositoryv1.Commit, int32, int32, error) {
	return []*coderepositoryv1.Commit{}, 0, 0, nil
}
func (f *fakeService) UpdateCommit(ctx context.Context, projectID, repositoryID, commitID uint64, message string) (*coderepositoryv1.Commit, error) {
	return nil, nil
}
func (f *fakeService) DeleteCommit(ctx context.Context, projectID, repositoryID, commitID uint64) error {
	return nil
}

func withUser(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{"x-user-id": "1"})
	return metadata.NewIncomingContext(ctx, md)
}

func TestListRepositories_PaginationValidation(t *testing.T) {
	h := NewCodeRepositoryGRPCHandler(&fakeService{})
	ctx := withUser(context.Background())

	// page < 1 (negative) -> error
	_, err := h.ListRepositories(ctx, &coderepositoryv1.ListRepositoriesRequest{ProjectId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize -> error
	_, err = h.ListRepositories(ctx, &coderepositoryv1.ListRepositoriesRequest{ProjectId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults when page=0 and page_size=0
	resp, err := h.ListRepositories(ctx, &coderepositoryv1.ListRepositoriesRequest{ProjectId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}

func TestGetRepositoryBranches_PaginationDefaultsAndValidation(t *testing.T) {
	h := NewCodeRepositoryGRPCHandler(&fakeService{})
	ctx := withUser(context.Background())

	// defaults
	resp, err := h.GetRepositoryBranches(ctx, &coderepositoryv1.GetRepositoryBranchesRequest{ProjectId: 1, RepositoryId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}

	// invalid page
	_, err = h.GetRepositoryBranches(ctx, &coderepositoryv1.GetRepositoryBranchesRequest{ProjectId: 1, RepositoryId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}
}

func TestListBranches_PaginationValidation(t *testing.T) {
	h := NewCodeRepositoryGRPCHandler(&fakeService{})
	ctx := withUser(context.Background())

	_, err := h.ListBranches(ctx, &coderepositoryv1.ListBranchesRequest{ProjectId: 1, RepositoryId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	_, err = h.ListBranches(ctx, &coderepositoryv1.ListBranchesRequest{ProjectId: 1, RepositoryId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}
}

func TestListCommits_PaginationValidation(t *testing.T) {
	h := NewCodeRepositoryGRPCHandler(&fakeService{})
	ctx := withUser(context.Background())

	_, err := h.ListCommits(ctx, &coderepositoryv1.ListCommitsRequest{ProjectId: 1, RepositoryId: 1, BranchId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	_, err = h.ListCommits(ctx, &coderepositoryv1.ListCommitsRequest{ProjectId: 1, RepositoryId: 1, BranchId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}
}
