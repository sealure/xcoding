package server

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"xcoding/apps/user/internal/service"
	userv1 "xcoding/gen/go/user/v1"
)

type UserGRPCHandler struct {
	userv1.UnimplementedUserServiceServer // UnimplementedUserServiceServer 为所有 gRPC 服务方法提供默认实现（返回“未实现”错误）
	userService                           service.UserService
}

func NewUserGRPCHandler(userService service.UserService) *UserGRPCHandler {
	return &UserGRPCHandler{
		userService: userService,
	}
}

func (s *UserGRPCHandler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	user, token, err := s.userService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserGRPCHandler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	user, token, err := s.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &userv1.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserGRPCHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	userID := req.GetUserId()

	user, err := s.userService.GetUser(ctx, &userID)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserResponse{
		User: user,
	}, nil
}

func (s *UserGRPCHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	userID := req.GetUserId()

	updates := make(map[string]interface{})
	if req.GetUsername() != "" {
		updates["username"] = req.GetUsername()
	}
	if req.GetEmail() != "" {
		updates["email"] = req.GetEmail()
	}
	if req.GetAvatar() != "" {
		updates["avatar"] = req.GetAvatar()
	}

	if req.ProtoReflect().Has(req.ProtoReflect().Descriptor().Fields().ByNumber(5)) {
		updates["role"] = req.GetRole()
	}

	if req.ProtoReflect().Has(req.ProtoReflect().Descriptor().Fields().ByNumber(6)) {
		updates["is_active"] = req.GetIsActive()
	}

	user, err := s.userService.UpdateUser(ctx, &userID, updates)
	if err != nil {
		return nil, err
	}

	return &userv1.UpdateUserResponse{
		User: user,
	}, nil
}

func (s *UserGRPCHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	page, pageSize, err := normalizePagination(req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}

	users, total, totalPages, err := s.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	pagination := &userv1.ListUsersResponse_Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: int32(total),
		TotalPages: int32(totalPages),
	}

	return &userv1.ListUsersResponse{
		Data:       users,
		Pagination: pagination,
	}, nil
}

func (s *UserGRPCHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &userv1.DeleteUserResponse{
		Success: true,
	}, nil
}

func (s *UserGRPCHandler) Auth(ctx context.Context, req *userv1.AuthRequest) (*userv1.AuthResponse, error) {
	// 从gRPC元数据中提取Authorization头和其他信息
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// 提取Authorization头中的token
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			// 正确处理Bearer前缀
			authHeader := authHeaders[0]
			if strings.HasPrefix(authHeader, "Bearer ") {
				req.Token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				req.Token = authHeader
			}
			req.TokenType = userv1.TokenType_TOKEN_TYPE_UNSPECIFIED // 让服务自动检测令牌类型
		}

		// 提取其他请求信息
		if methodHeaders := md.Get("x-request-method"); len(methodHeaders) > 0 {
			req.HttpMethod = methodHeaders[0]
		}
		if pathHeaders := md.Get("x-request-path"); len(pathHeaders) > 0 {
			req.RequestPath = pathHeaders[0]
		}
		if ipHeaders := md.Get("x-client-ip"); len(ipHeaders) > 0 {
			req.ClientIp = ipHeaders[0]
		}
		if uaHeaders := md.Get("x-user-agent"); len(uaHeaders) > 0 {
			req.UserAgent = uaHeaders[0]
		}
	}

	resp, err := s.userService.Auth(ctx, req.Token, req.TokenType, req.HttpMethod, req.RequestPath, req.ClientIp, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserGRPCHandler) CreateAPIToken(ctx context.Context, req *userv1.CreateAPITokenRequest) (*userv1.CreateAPITokenResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	resp, err := s.userService.CreateAPIToken(ctx, req.Name, req.Description, req.ExpiresIn, req.Scopes)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserGRPCHandler) ListAPITokens(ctx context.Context, req *userv1.ListAPITokensRequest) (*userv1.ListAPITokensResponse, error) {
	tokens, err := s.userService.ListAPITokens(ctx)
	if err != nil {
		return nil, err
	}

	return &userv1.ListAPITokensResponse{
		Tokens: tokens,
	}, nil
}

func (s *UserGRPCHandler) DeleteAPIToken(ctx context.Context, req *userv1.DeleteAPITokenRequest) (*userv1.DeleteAPITokenResponse, error) {
	if req.GetTokenId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "token ID is required")
	}

	err := s.userService.DeleteAPIToken(ctx, req.GetTokenId())
	if err != nil {
		return nil, err
	}

	return &userv1.DeleteAPITokenResponse{
		Success: true,
	}, nil
}
