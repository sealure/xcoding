package service

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"xcoding/apps/project/internal/models"
	projectv1 "xcoding/gen/go/project/v1"
)

// normalizeRole trims and uppercases a role string for reliable comparison.
func normalizeRole(role string) string { return strings.ToUpper(strings.TrimSpace(role)) }

// isUserRoleSuperAdmin checks whether the incoming request has global super admin role
// based on X-User-Role header emitted by user.Auth (enum string: USER_ROLE_SUPER_ADMIN).
func isUserRoleSuperAdmin(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	vals := md.Get("x-user-role")
	if len(vals) == 0 {
		return false
	}
	r := normalizeRole(vals[0])
	return r == normalizeRole("USER_ROLE_SUPER_ADMIN")
}

// isProjectMemberRoleMember determines whether actor is a MEMBER in given project.
func (s *projectService) isProjectMemberRoleMember(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	var mem models.ProjectMember
	if err := s.db.WithContext(ctx).First(&mem, &models.ProjectMember{ProjectID: projectID, UserID: actorID}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, status.Errorf(codes.Internal, "failed to check member: %v", err)
	}
	return mem.Role == int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_MEMBER), nil
}

// isProjectMemberRoleAdminOrOwner determines whether actor is ADMIN or OWNER of given project.
func (s *projectService) isProjectMemberRoleAdminOrOwner(ctx context.Context, projectID uint64, actorID uint64) (bool, error) {
	var p models.Project
	if err := s.db.WithContext(ctx).First(&p, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, status.Errorf(codes.NotFound, "project not found")
		}
		return false, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}
	if p.OwnerID == actorID {
		return true, nil
	}
	var mem models.ProjectMember
	if err := s.db.WithContext(ctx).First(&mem, &models.ProjectMember{ProjectID: projectID, UserID: actorID}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, status.Errorf(codes.Internal, "failed to check member: %v", err)
	}
	return mem.Role == int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_ADMIN) || mem.Role == int32(projectv1.ProjectMemberRole_PROJECT_MEMBER_ROLE_OWNER), nil
}

// isGlobalAdmin delegates to user-role super admin check.
// Project-level ADMIN/OWNER checks should be handled by call sites using isProjectMemberRoleAdminOrOwner.
func isGlobalAdmin(ctx context.Context) bool { return isUserRoleSuperAdmin(ctx) }
