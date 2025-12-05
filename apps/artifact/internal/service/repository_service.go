package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/artifact/internal/models"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 制品仓库（Repository）操作
func (s *artifactService) CreateRepository(ctx context.Context, name, description string, namespaceID uint64, isPublic bool, path string) (*artifactv1.Repository, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	var existingRepository models.Repository
	err = s.db.WithContext(ctx).Where("name = ? AND namespace_id = ?", name, namespaceID).First(&existingRepository).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "repository already exists in this namespace")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check repository: %v", err)
	}

	// 检查命名空间是否存在
	var namespace models.Namespace
	err = s.db.WithContext(ctx).Preload("Registry").First(&namespace, namespaceID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "namespace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}

	projectID := namespace.Registry.ProjectID
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return nil, err
		}
	}

	repository := models.Repository{
		Name:        name,
		Description: description,
		NamespaceID: namespaceID,
		IsPublic:    isPublic,
		Path:        path,
	}

	// 显式选择字段以确保 is_public=false 被持久化，而不是使用 DB 默认值
	if err := s.db.WithContext(ctx).
		Select("Name", "Description", "NamespaceID", "IsPublic", "Path").
		Create(&repository).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create repository: %v", err)
	}

	return repository.ToProto(), nil
}

func (s *artifactService) GetRepository(ctx context.Context, id uint64) (*artifactv1.Repository, error) {
	var repository models.Repository
	if err := s.db.WithContext(ctx).Preload("Namespace").Preload("Namespace.Registry").First(&repository, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	// 非超级管理员访问控制：私有资源需成员身份才能访问
	if !isUserRoleSuperAdmin(ctx) && !repository.IsPublic && !repository.Namespace.Registry.IsPublic {
		actorID, aerr := getUserIDFromCtx(ctx)
		if aerr != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", aerr)
		}
		projectID := repository.Namespace.Registry.ProjectID
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to access private repository")
		}
	}

	return repository.ToProtoWithNamespace(), nil
}

func (s *artifactService) UpdateRepository(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Repository, error) {
	var repository models.Repository
	if err := s.db.WithContext(ctx).First(&repository, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}

	if name, ok := updates["name"].(string); ok {
		repository.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		repository.Description = description
	}
	if namespaceID, ok := updates["namespace_id"].(uint64); ok {
		repository.NamespaceID = namespaceID
	}
	// 支持将 is_public 设置为 false，不用依赖 truthy
	if v, ok := updates["is_public"]; ok {
		if b, bok := v.(bool); bok {
			repository.IsPublic = b
		}
	}
	if path, ok := updates["path"].(string); ok {
		repository.Path = path
	}

	if err := s.db.WithContext(ctx).Save(&repository).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update repository: %v", err)
	}

	return s.GetRepository(ctx, id)
}

func (s *artifactService) ListRepositories(ctx context.Context, page, pageSize int32, namespaceID uint64) ([]*artifactv1.Repository, int32, int32, error) {
	offset := (page - 1) * pageSize

	var repositories []models.Repository
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Repository{})
	if namespaceID > 0 {
		query = query.Where("namespace_id = ?", namespaceID)
	}
	if !isUserRoleSuperAdmin(ctx) {
		query = query.Joins("JOIN namespaces ON repositories.namespace_id = namespaces.id").Joins("JOIN registries ON namespaces.registry_id = registries.id").Where("repositories.is_public = ? OR registries.is_public = ?", true, true)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count repositories: %v", err)
	}
	if err := query.Preload("Namespace").Preload("Namespace.Registry").Offset(int(offset)).Limit(int(pageSize)).Find(&repositories).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list repositories: %v", err)
	}

	repositoryProtos := make([]*artifactv1.Repository, len(repositories))
	for i, repository := range repositories {
		repositoryProtos[i] = repository.ToProtoWithNamespace()
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return repositoryProtos, int32(total), totalPages, nil
}

func (s *artifactService) DeleteRepository(ctx context.Context, id uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	var repository models.Repository
	if err := s.db.WithContext(ctx).Preload("Namespace").Preload("Namespace.Registry").First(&repository, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "repository not found")
		}
		return status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	projectID := repository.Namespace.Registry.ProjectID
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return err
		}
	}

	if err := s.db.WithContext(ctx).Delete(&repository).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete repository: %v", err)
	}

	return nil
}
