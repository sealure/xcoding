package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	artifactv1 "xcoding/gen/go/artifact/v1"
	projectv1 "xcoding/gen/go/project/v1"
	"xcoding/pkg/auth"
)

type ArtifactService interface {
	// 注册中心（Registry）相关操作
	CreateRegistry(ctx context.Context, name, url, description, username, password string, isPublic bool, projectID uint64, artifactType artifactv1.ArtifactType, artifactSource artifactv1.ArtifactSource) (*artifactv1.Registry, error)
	GetRegistry(ctx context.Context, id uint64, projectID uint64) (*artifactv1.Registry, error)
	UpdateRegistry(ctx context.Context, id uint64, updates map[string]interface{}, projectID uint64) (*artifactv1.Registry, error)
	ListRegistries(ctx context.Context, page, pageSize int32, projectID uint64) ([]*artifactv1.Registry, int32, int32, error)
	DeleteRegistry(ctx context.Context, id uint64, projectID uint64) error

	// 命名空间（Namespace）相关操作
	CreateNamespace(ctx context.Context, name, description string, registryID uint64) (*artifactv1.Namespace, error)
	GetNamespace(ctx context.Context, id uint64) (*artifactv1.Namespace, error)
	UpdateNamespace(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Namespace, error)
	ListNamespaces(ctx context.Context, page, pageSize int32, registryID uint64) ([]*artifactv1.Namespace, int32, int32, error)
	DeleteNamespace(ctx context.Context, id uint64) error

	// 制品仓库（Repository）相关操作
	CreateRepository(ctx context.Context, name, description string, namespaceID uint64, isPublic bool, path string) (*artifactv1.Repository, error)
	GetRepository(ctx context.Context, id uint64) (*artifactv1.Repository, error)
	UpdateRepository(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Repository, error)
	ListRepositories(ctx context.Context, page, pageSize int32, namespaceID uint64) ([]*artifactv1.Repository, int32, int32, error)
	DeleteRepository(ctx context.Context, id uint64) error

	// 制品标签（Tag）相关操作
	CreateTag(ctx context.Context, name, digest string, size int64, repositoryID uint64, isLatest bool) (*artifactv1.Tag, error)
	GetTag(ctx context.Context, id uint64) (*artifactv1.Tag, error)
	UpdateTag(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Tag, error)
	ListTags(ctx context.Context, page, pageSize int32, repositoryID uint64) ([]*artifactv1.Tag, int32, int32, error)
	DeleteTag(ctx context.Context, id uint64) error
}

type artifactService struct {
	db            *gorm.DB
	projectClient projectv1.ProjectServiceClient
}

func NewArtifactService(db *gorm.DB, projectClient projectv1.ProjectServiceClient) ArtifactService {
	return &artifactService{
		db:            db,
		projectClient: projectClient,
	}
}

// ==== 权限相关辅助函数 ====
func getUserIDFromCtx(ctx context.Context) (uint64, error) { return auth.GetUserIDFromCtx(ctx) }
func isUserRoleSuperAdmin(ctx context.Context) bool        { return auth.IsUserRoleSuperAdmin(ctx) }

func (s *artifactService) isMemberOrHigher(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	p := resp.GetProject()
	if p == nil {
		return false, status.Errorf(codes.NotFound, "project not found")
	}
	if p.OwnerId == actorID {
		return true, nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			return true, nil
		}
	}
	return false, nil
}

func (s *artifactService) ensureOwnerOrAdmin(ctx context.Context, projectID uint64, actorID uint64) error {
	if isUserRoleSuperAdmin(ctx) {
		return nil
	}
	resp, err := s.projectClient.GetProject(ctx, &projectv1.GetProjectRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	p := resp.GetProject()
	if p == nil {
		return status.Errorf(codes.NotFound, "project not found")
	}
	if p.OwnerId == actorID {
		return nil
	}
	members, err := s.projectClient.ListProjectMembers(ctx, &projectv1.ListProjectMembersRequest{ProjectId: projectID})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list project members: %v", err)
	}
	for _, m := range members.GetData() {
		if m.GetUserId() == actorID {
			role := m.GetRole()
			if role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER || role == projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN {
				return nil
			}
		}
	}
	return status.Errorf(codes.PermissionDenied, "only owner or admin can perform this action")
}
