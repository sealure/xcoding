package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/artifact/internal/models"
	artifactv1 "xcoding/gen/go/artifact/v1"
)

// 命名空间（Namespace）操作
func (s *artifactService) CreateNamespace(ctx context.Context, name, description string, registryID uint64) (*artifactv1.Namespace, error) {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can create namespace")
	}
	var existingNamespace models.Namespace
	err := s.db.WithContext(ctx).Where("name = ? AND registry_id = ?", name, registryID).First(&existingNamespace).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "namespace already exists in this registry")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check namespace: %v", err)
	}

	// 检查注册中心是否存在
	var registry models.Registry
	err = s.db.WithContext(ctx).First(&registry, registryID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "registry not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get registry: %v", err)
	}

	namespace := models.Namespace{
		Name:        name,
		Description: description,
		RegistryID:  registryID,
	}

	if err := s.db.WithContext(ctx).Create(&namespace).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create namespace: %v", err)
	}

	return namespace.ToProto(), nil
}

func (s *artifactService) GetNamespace(ctx context.Context, id uint64) (*artifactv1.Namespace, error) {
	var namespace models.Namespace
	err := s.db.WithContext(ctx).Preload("Registry").First(&namespace, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "namespace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}
	// 非超级管理员仅可访问公共注册中心下的命名空间
	if !isUserRoleSuperAdmin(ctx) && !namespace.Registry.IsPublic {
		return nil, status.Errorf(codes.PermissionDenied, "not allowed to access private namespace")
	}
	return namespace.ToProtoWithRegistry(), nil
}

func (s *artifactService) UpdateNamespace(ctx context.Context, id uint64, updates map[string]interface{}) (*artifactv1.Namespace, error) {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can update namespace")
	}
	var namespace models.Namespace
	if err := s.db.WithContext(ctx).First(&namespace, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "namespace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}

	if name, ok := updates["name"].(string); ok {
		namespace.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		namespace.Description = description
	}

	if err := s.db.WithContext(ctx).Save(&namespace).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update namespace: %v", err)
	}

	return s.GetNamespace(ctx, id)
}

func (s *artifactService) ListNamespaces(ctx context.Context, page, pageSize int32, registryID uint64) ([]*artifactv1.Namespace, int32, int32, error) {
	offset := (page - 1) * pageSize

	var namespaces []models.Namespace
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Namespace{})
	if registryID > 0 {
		query = query.Where("registry_id = ?", registryID)
	}
	if !isUserRoleSuperAdmin(ctx) {
		query = query.Joins("JOIN registries ON namespaces.registry_id = registries.id").Where("registries.is_public = ?", true)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count namespaces: %v", err)
	}
	if err := query.Preload("Registry").Offset(int(offset)).Limit(int(pageSize)).Find(&namespaces).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list namespaces: %v", err)
	}

	namespaceProtos := make([]*artifactv1.Namespace, len(namespaces))
	for i, namespace := range namespaces {
		namespaceProtos[i] = namespace.ToProtoWithRegistry()
	}

	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	return namespaceProtos, int32(total), totalPages, nil
}

func (s *artifactService) DeleteNamespace(ctx context.Context, id uint64) error {
	if _, err := getUserIDFromCtx(ctx); err != nil {
		return status.Errorf(codes.Unauthenticated, "missing user: %v", err)
	}
	if !isUserRoleSuperAdmin(ctx) {
		return status.Errorf(codes.PermissionDenied, "only super admins can delete namespace")
	}
	var namespace models.Namespace
	if err := s.db.WithContext(ctx).First(&namespace, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return status.Errorf(codes.NotFound, "namespace not found")
		}
		return status.Errorf(codes.Internal, "failed to get namespace: %v", err)
	}

	if err := s.db.WithContext(ctx).Delete(&namespace).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete namespace: %v", err)
	}

	return nil
}
