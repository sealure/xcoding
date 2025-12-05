package service

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
	userv1 "xcoding/gen/go/user/v1"
)

// isUserRoleSuperAdmin 判断当前请求是否为全局超级管理员
// 依据网关写入的 X-User-Role 头，值来源于 user.Auth 返回的枚举字符串
// 允许大小写不敏感兼容 "SUPER_ADMIN" 简写
func isUserRoleSuperAdmin(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	vals := md.Get("x-user-role")
	if len(vals) == 0 {
		return false
	}
	role := strings.TrimSpace(vals[0])
	return strings.EqualFold(role, userv1.UserRole_USER_ROLE_SUPER_ADMIN.String()) ||
		strings.EqualFold(role, "SUPER_ADMIN")
}
