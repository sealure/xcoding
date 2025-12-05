package handler

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xcoding/apps/artifact/internal/service"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// fake service implementing service.ArtifactService
type fakeArtifactService struct{}

// Registry operations
func (f *fakeArtifactService) CreateRegistry(ctx context.Context, name, url, description, username, password string, isPublic bool, projectID uint64, artifactType artifactv1.ArtifactType, artifactSource artifactv1.ArtifactSource) (*artifactv1.Registry, error) {
	return nil, nil
}
func (f *fakeArtifactService) GetRegistry(ctx context.Context, id uint64, projectID uint64) (*artifactv1.Registry, error) {
	return nil, nil
}
func (f *fakeArtifactService) UpdateRegistry(ctx context.Context, id uint64, updates map[string]interface{}, projectID uint64) (*artifactv1.Registry, error) {
	return nil, nil
}
func (f *fakeArtifactService) ListRegistries(ctx context.Context, page, pageSize int32, projectID uint64) ([]*artifactv1.Registry, int32, int32, error) {
	return []*artifactv1.Registry{}, 0, 0, nil
}
func (f *fakeArtifactService) DeleteRegistry(ctx context.Context, id uint64, projectID uint64) error {
	return nil
}

// Namespace operations
func (f *fakeArtifactService) CreateNamespace(ctx context.Context, name, description string, registryID uint64) (*artifactv1.Namespace, error) {
	return nil, nil
}
func (f *fakeArtifactService) GetNamespace(ctx context.Context, id uint64) (*artifactv1.Namespace, error) {
	return nil, nil
}
func (f *fakeArtifactService) UpdateNamespace(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Namespace, error) {
	return nil, nil
}
func (f *fakeArtifactService) ListNamespaces(ctx context.Context, page, pageSize int32, registryID uint64) ([]*artifactv1.Namespace, int32, int32, error) {
	return []*artifactv1.Namespace{}, 0, 0, nil
}
func (f *fakeArtifactService) DeleteNamespace(ctx context.Context, id uint64) error { return nil }

// Repository operations
func (f *fakeArtifactService) CreateRepository(ctx context.Context, name, description string, namespaceID uint64, isPublic bool, path string) (*artifactv1.Repository, error) {
	return nil, nil
}
func (f *fakeArtifactService) GetRepository(ctx context.Context, id uint64) (*artifactv1.Repository, error) {
	return nil, nil
}
func (f *fakeArtifactService) UpdateRepository(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Repository, error) {
	return nil, nil
}
func (f *fakeArtifactService) ListRepositories(ctx context.Context, page, pageSize int32, namespaceID uint64) ([]*artifactv1.Repository, int32, int32, error) {
	return []*artifactv1.Repository{}, 0, 0, nil
}
func (f *fakeArtifactService) DeleteRepository(ctx context.Context, id uint64) error { return nil }

// Tag operations
func (f *fakeArtifactService) CreateTag(ctx context.Context, name, digest string, size int64, repositoryID uint64, isLatest bool) (*artifactv1.Tag, error) {
	return nil, nil
}
func (f *fakeArtifactService) GetTag(ctx context.Context, id uint64) (*artifactv1.Tag, error) {
	return nil, nil
}
func (f *fakeArtifactService) UpdateTag(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Tag, error) {
	return nil, nil
}
func (f *fakeArtifactService) ListTags(ctx context.Context, page, pageSize int32, repositoryID uint64) ([]*artifactv1.Tag, int32, int32, error) {
	return []*artifactv1.Tag{}, 0, 0, nil
}
func (f *fakeArtifactService) DeleteTag(ctx context.Context, id uint64) error { return nil }

// Image operations
func (f *fakeArtifactService) GetImageLayers(ctx context.Context, registryID, namespaceID, repositoryID uint64, tagName string) ([]string, error) {
	return []string{}, nil
}
func (f *fakeArtifactService) DeleteImage(ctx context.Context, registryID, namespaceID, repositoryID uint64, tagName string) error {
	return nil
}

var _ service.ArtifactService = (*fakeArtifactService)(nil)

func TestListRegistries_PaginationValidation(t *testing.T) {
	h := NewArtifactGRPCHandler(&fakeArtifactService{})
	ctx := context.Background()

	// invalid page
	_, err := h.ListRegistries(ctx, &artifactv1.ListRegistriesRequest{Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}

	// page_size > maxPageSize
	_, err = h.ListRegistries(ctx, &artifactv1.ListRegistriesRequest{Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListRegistries(ctx, &artifactv1.ListRegistriesRequest{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}

func TestListNamespaces_PaginationValidation(t *testing.T) {
	h := NewArtifactGRPCHandler(&fakeArtifactService{})
	ctx := context.Background()

	// invalid page
	_, err := h.ListNamespaces(ctx, &artifactv1.ListNamespacesRequest{RegistryId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize
	_, err = h.ListNamespaces(ctx, &artifactv1.ListNamespacesRequest{RegistryId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListNamespaces(ctx, &artifactv1.ListNamespacesRequest{RegistryId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}

func TestListRepositories_PaginationValidation(t *testing.T) {
	h := NewArtifactGRPCHandler(&fakeArtifactService{})
	ctx := context.Background()

	// invalid page
	_, err := h.ListRepositories(ctx, &artifactv1.ListRepositoriesRequest{NamespaceId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize
	_, err = h.ListRepositories(ctx, &artifactv1.ListRepositoriesRequest{NamespaceId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListRepositories(ctx, &artifactv1.ListRepositoriesRequest{NamespaceId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}

func TestListTags_PaginationValidation(t *testing.T) {
	h := NewArtifactGRPCHandler(&fakeArtifactService{})
	ctx := context.Background()

	// invalid page
	_, err := h.ListTags(ctx, &artifactv1.ListTagsRequest{RepositoryId: 1, Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size > maxPageSize
	_, err = h.ListTags(ctx, &artifactv1.ListTagsRequest{RepositoryId: 1, Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// defaults
	resp, err := h.ListTags(ctx, &artifactv1.ListTagsRequest{RepositoryId: 1, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}
