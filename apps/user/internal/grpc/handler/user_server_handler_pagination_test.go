package server

import (
	"context"
	"testing"

	userv1 "xcoding/gen/go/user/v1"
)

// UserService 的伪实现（用于测试）
type fakeUserService struct{}

func (f *fakeUserService) Register(ctx context.Context, username, email, password string) (*userv1.User, string, error) {
	return nil, "", nil
}
func (f *fakeUserService) Login(ctx context.Context, username, password string) (*userv1.User, string, error) {
	return nil, "", nil
}
func (f *fakeUserService) GetUser(ctx context.Context, userID *uint64) (*userv1.User, error) {
	return nil, nil
}
func (f *fakeUserService) UpdateUser(ctx context.Context, userID *uint64, updates map[string]interface{}) (*userv1.User, error) {
	return nil, nil
}
func (f *fakeUserService) ListUsers(ctx context.Context, page, pageSize int32) ([]*userv1.User, int32, int32, error) {
	return []*userv1.User{}, 0, 0, nil
}
func (f *fakeUserService) DeleteUser(ctx context.Context, id uint64) error { return nil }
func (f *fakeUserService) Auth(ctx context.Context, token string, tokenType userv1.TokenType, httpMethod, requestPath, clientIP, userAgent string) (*userv1.AuthResponse, error) {
	return &userv1.AuthResponse{Authenticated: true}, nil
}
func (f *fakeUserService) CreateAPIToken(ctx context.Context, name, description string, expiresIn userv1.TokenExpiration, scopes []userv1.Scope) (*userv1.CreateAPITokenResponse, error) {
	return nil, nil
}
func (f *fakeUserService) ListAPITokens(ctx context.Context) ([]*userv1.CreateAPITokenResponse, error) {
	return []*userv1.CreateAPITokenResponse{}, nil
}
func (f *fakeUserService) DeleteAPIToken(ctx context.Context, tokenID uint64) error { return nil }

func TestListUsers_PaginationValidation(t *testing.T) {
	h := NewUserGRPCHandler(&fakeUserService{})
	ctx := context.Background()

	// 非法的 page 值
	_, err := h.ListUsers(ctx, &userv1.ListUsersRequest{Page: -1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected error for page=-1")
	}

	// page_size 大于最大允许值（maxPageSize）
	_, err = h.ListUsers(ctx, &userv1.ListUsersRequest{Page: 1, PageSize: maxPageSize + 1})
	if err == nil {
		t.Fatalf("expected error for page_size > maxPageSize")
	}

	// 默认值校验
	resp, err := h.ListUsers(ctx, &userv1.ListUsersRequest{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected error for defaults: %v", err)
	}
	if resp.Pagination.GetPage() != 1 || resp.Pagination.GetPageSize() != 10 {
		t.Fatalf("expected defaults page=1, page_size=10, got page=%d, size=%d", resp.Pagination.GetPage(), resp.Pagination.GetPageSize())
	}
}
