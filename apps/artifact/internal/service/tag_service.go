package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/artifact/internal/models"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 制品标签（Tag）操作
func (s *artifactService) CreateTag(ctx context.Context, name, digest string, size int64, repositoryID uint64, isLatest bool) (*artifactv1.Tag, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}

	// 确认仓库存在，并基于项目所有者身份校验权限
	var repository models.Repository
	if err := s.db.WithContext(ctx).Preload("Namespace").Preload("Namespace.Registry").First(&repository, repositoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "repository not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get repository: %v", err)
	}
	projectID := repository.Namespace.Registry.ProjectID
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return nil, err
		}
	}

	// 保证标签名在同一仓库内唯一
	var existingTag models.Tag
	if err := s.db.WithContext(ctx).Where("name = ? AND repository_id = ?", name, repositoryID).First(&existingTag).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "tag already exists in this repository")
	} else if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check tag: %v", err)
	}

	newTag := models.Tag{RepositoryID: repositoryID, Name: name, Digest: digest, SizeBytes: size, IsLatest: isLatest}
	if err := s.db.WithContext(ctx).Create(&newTag).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
	}
	return newTag.ToProto(), nil
}

func (s *artifactService) GetTag(ctx context.Context, id uint64) (*artifactv1.Tag, error) {
	var tag models.Tag
	if err := s.db.WithContext(ctx).Preload("Repository").Preload("Repository.Namespace").Preload("Repository.Namespace.Registry").First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}
	// 非超级管理员访问控制：私有资源不开放
	if !isUserRoleSuperAdmin(ctx) && !tag.Repository.IsPublic && !tag.Repository.Namespace.Registry.IsPublic {
		actorID, aerr := getUserIDFromCtx(ctx)
		if aerr != nil {
			return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", aerr)
		}
		projectID := tag.Repository.Namespace.Registry.ProjectID
		ok, perr := s.isMemberOrHigher(ctx, projectID, actorID)
		if perr != nil {
			return nil, perr
		}
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to access private tag")
		}
	}
	return tag.ToProtoWithRepository(), nil
}

func (s *artifactService) UpdateTag(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Tag, error) {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	var tag models.Tag
	if err := s.db.WithContext(ctx).Preload("Repository").Preload("Repository.Namespace").Preload("Repository.Namespace.Registry").First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "tag not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}
	projectID := tag.Repository.Namespace.Registry.ProjectID
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return nil, err
		}
	}
	if name, ok := updates["name"].(string); ok {
		tag.Name = name
	}
	if digest, ok := updates["digest"].(string); ok {
		tag.Digest = digest
	}
	if size, ok := updates["size"].(int64); ok {
		tag.SizeBytes = size
	}
	if isLatest, ok := updates["is_latest"].(bool); ok {
		tag.IsLatest = isLatest
	}
	if err := s.db.WithContext(ctx).Save(&tag).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
	}
	return s.GetTag(ctx, id)
}

func (s *artifactService) ListTags(ctx context.Context, page, pageSize int32, repositoryID uint64) ([]*artifactv1.Tag, int32, int32, error) {
	offset := (page - 1) * pageSize
	var tags []models.Tag
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Tag{})
	if repositoryID > 0 {
		query = query.Where("repository_id = ?", repositoryID)
	}
	if !isUserRoleSuperAdmin(ctx) {
		query = query.Joins("JOIN repositories ON tags.repository_id = repositories.id").Joins("JOIN namespaces ON repositories.namespace_id = namespaces.id").Joins("JOIN registries ON namespaces.registry_id = registries.id").Where("repositories.is_public = ? OR registries.is_public = ?", true, true)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count tags: %v", err)
	}
	if err := query.Preload("Repository").Preload("Repository.Namespace").Preload("Repository.Namespace.Registry").Offset(int(offset)).Limit(int(pageSize)).Find(&tags).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list tags: %v", err)
	}

	protos := make([]*artifactv1.Tag, len(tags))
	for i, tag := range tags {
		protos[i] = tag.ToProtoWithRepository()
	}
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return protos, int32(total), totalPages, nil
}

func (s *artifactService) DeleteTag(ctx context.Context, id uint64) error {
	actorID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	var tag models.Tag
	if err := s.db.WithContext(ctx).Preload("Repository").Preload("Repository.Namespace").Preload("Repository.Namespace.Registry").First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "tag not found")
		}
		return status.Errorf(codes.Internal, "failed to get tag: %v", err)
	}
	projectID := tag.Repository.Namespace.Registry.ProjectID
	if !isUserRoleSuperAdmin(ctx) {
		if err := s.ensureOwnerOrAdmin(ctx, projectID, actorID); err != nil {
			return err
		}
	}
	if err := s.db.WithContext(ctx).Delete(&tag).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete tag: %v", err)
	}
	return nil
}
