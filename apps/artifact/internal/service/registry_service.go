package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/artifact/internal/models"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 注册中心（Registry）相关操作
func (s *artifactService) CreateRegistry(ctx context.Context, name, url, description, username, password string, isPublic bool, projectID uint64, artifactType artifactv1.ArtifactType, artifactSource artifactv1.ArtifactSource) (*artifactv1.Registry, error) {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can create registry")
	}
	var existingRegistry models.Registry
	err := s.db.WithContext(ctx).Where("name = ?", name).First(&existingRegistry).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "registry already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check registry: %v", err)
	}

	registry := models.Registry{
		Name:           name,
		URL:            url,
		Description:    description,
		IsPublic:       isPublic,
		Username:       username,
		Password:       password,
		ProjectID:      projectID,
		ArtifactType:   models.ArtifactTypeFromProto(artifactType),
		ArtifactSource: models.ArtifactSourceFromProto(artifactSource),
	}

	// 显式选择字段以确保 is_public=false 被持久化，而不是使用 DB 默认值
	if err := s.db.WithContext(ctx).
		Select("Name", "URL", "Description", "IsPublic", "Username", "Password", "ProjectID", "ArtifactType", "ArtifactSource").
		Create(&registry).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registry: %v", err)
	}

	return registry.ToProto(), nil
}

func (s *artifactService) GetRegistry(ctx context.Context, id uint64, projectID uint64) (*artifactv1.Registry, error) {
	var registry models.Registry
	err := s.db.WithContext(ctx).First(&registry, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "registry not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get registry: %v", err)
	}
	if projectID > 0 && registry.ProjectID != projectID {
		return nil, status.Errorf(codes.PermissionDenied, "registry does not belong to specified project")
	}
	// 非超级管理员仅可访问公共注册中心
	if !isUserRoleSuperAdmin(ctx) && !registry.IsPublic {
		return nil, status.Errorf(codes.PermissionDenied, "not allowed to access private registry")
	}
	return registry.ToProto(), nil
}

func (s *artifactService) UpdateRegistry(ctx context.Context, id uint64, updates map[string]interface{}, projectID uint64) (*artifactv1.Registry, error) {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can update registry")
	}
	var registry models.Registry
	if err := s.db.WithContext(ctx).First(&registry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "registry not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get registry: %v", err)
	}
	if projectID > 0 && registry.ProjectID != projectID {
		return nil, status.Errorf(codes.PermissionDenied, "registry does not belong to specified project")
	}

	if name, ok := updates["name"].(string); ok {
		registry.Name = name
	}
	if url, ok := updates["url"].(string); ok {
		registry.URL = url
	}
	if description, ok := updates["description"].(string); ok {
		registry.Description = description
	}
	if isPublic, ok := updates["is_public"].(bool); ok {
		registry.IsPublic = isPublic
	}
	if username, ok := updates["username"].(string); ok {
		registry.Username = username
	}
	if password, ok := updates["password"].(string); ok {
		registry.Password = password
	}
	if v, ok := updates["artifact_type"]; ok {
		if iv, ok2 := v.(int32); ok2 {
			registry.ArtifactType = models.ArtifactType(iv)
		}
	}
	if v, ok := updates["artifact_source"]; ok {
		if iv, ok2 := v.(int32); ok2 {
			registry.ArtifactSource = models.ArtifactSource(iv)
		}
	}

	if err := s.db.WithContext(ctx).Save(&registry).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update registry: %v", err)
	}

	return s.GetRegistry(ctx, id, projectID)
}

func (s *artifactService) ListRegistries(ctx context.Context, page, pageSize int32, projectID uint64) ([]*artifactv1.Registry, int32, int32, error) {
	offset := (page - 1) * pageSize

	var registries []models.Registry
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Registry{})
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if !isUserRoleSuperAdmin(ctx) {
		query = query.Where("is_public = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count registries: %v", err)
	}
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&registries).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list registries: %v", err)
	}

	registryProtos := make([]*artifactv1.Registry, len(registries))
	for i, registry := range registries {
		registryProtos[i] = registry.ToProto()
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return registryProtos, int32(total), totalPages, nil
}

func (s *artifactService) DeleteRegistry(ctx context.Context, id uint64, projectID uint64) error {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return status.Errorf(codes.PermissionDenied, "only super admins can delete registry")
	}
	var registry models.Registry
	if err := s.db.WithContext(ctx).First(&registry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "registry not found")
		}
		return status.Errorf(codes.Internal, "failed to get registry: %v", err)
	}
	if projectID > 0 && registry.ProjectID != projectID {
		return status.Errorf(codes.PermissionDenied, "registry does not belong to specified project")
	}

	if err := s.db.WithContext(ctx).Delete(&registry).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete registry: %v", err)
	}

	return nil
}
