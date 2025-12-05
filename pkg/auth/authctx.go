package auth

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GetUserIDFromCtx extracts `x-user-id` from gRPC metadata and returns it as uint64.
// Returns gRPC status errors for uniform error handling across services.
func GetUserIDFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-user-id")
	if len(vals) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing x-user-id header")
	}
	id, err := strconv.ParseUint(strings.TrimSpace(vals[0]), 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid x-user-id: %v", err)
	}
	return id, nil
}

// GetUsernameFromCtx extracts `x-username` from gRPC metadata and returns it as string.
// Returns gRPC status errors for uniform error handling across services.
func GetUsernameFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-username")
	if len(vals) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing x-username header")
	}
	return strings.TrimSpace(vals[0]), nil
}

// GetProjectIDFromCtx extracts `x-project-id` from gRPC metadata and returns it as uint64.
// Returns 0 and nil if not present, allowing services to treat it as optional.
func GetProjectIDFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-project-id")
	if len(vals) == 0 {
		return 0, nil
	}
	id, err := strconv.ParseUint(strings.TrimSpace(vals[0]), 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "invalid x-project-id: %v", err)
	}
	return id, nil
}

// IsUserRoleSuperAdmin returns true if `x-user-role` indicates a super admin.
// Accepts both "USER_ROLE_SUPER_ADMIN" and case-insensitive "SUPER_ADMIN" values.
func IsUserRoleSuperAdmin(ctx context.Context) bool {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get("x-user-role")
		if len(vals) > 0 {
			role := strings.TrimSpace(vals[0])
			if role == "USER_ROLE_SUPER_ADMIN" || strings.EqualFold(role, "SUPER_ADMIN") {
				return true
			}
		}
	}
	return false
}

// MustSuperAdmin enforces super admin role and returns PermissionDenied if not.
func MustSuperAdmin(ctx context.Context) error {
	if !IsUserRoleSuperAdmin(ctx) {
		return status.Errorf(codes.PermissionDenied, "only super admins allowed")
	}
	return nil
}