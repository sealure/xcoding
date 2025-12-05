package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"xcoding/apps/user/internal/models"
	userv1 "xcoding/gen/go/user/v1"
)

type UserService interface {
	Register(ctx context.Context, username, email, password string) (*userv1.User, string, error)
	Login(ctx context.Context, username, password string) (*userv1.User, string, error)
	GetUser(ctx context.Context, userID *uint64) (*userv1.User, error)
	UpdateUser(ctx context.Context, userID *uint64, updates map[string]interface{}) (*userv1.User, error)
	ListUsers(ctx context.Context, page, pageSize int32) ([]*userv1.User, int32, int32, error)
	DeleteUser(ctx context.Context, id uint64) error
	Auth(ctx context.Context, token string, tokenType userv1.TokenType, httpMethod, requestPath, clientIP, userAgent string) (*userv1.AuthResponse, error)
	CreateAPIToken(ctx context.Context, name, description string, expiresIn userv1.TokenExpiration, scopes []userv1.Scope) (*userv1.CreateAPITokenResponse, error)
	ListAPITokens(ctx context.Context) ([]*userv1.CreateAPITokenResponse, error)
	DeleteAPIToken(ctx context.Context, tokenID uint64) error
}

type userService struct {
	db        *gorm.DB
	jwtSecret string
}

func NewUserService(db *gorm.DB, jwtSecret string) UserService {
	return &userService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *userService) Register(ctx context.Context, username, email, password string) (*userv1.User, string, error) {
	var existingUser models.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		return nil, "", status.Errorf(codes.AlreadyExists, "username already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, "", status.Errorf(codes.Internal, "failed to check username: %v", err)
	}

	err = s.db.WithContext(ctx).Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		return nil, "", status.Errorf(codes.AlreadyExists, "email already exists")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, "", status.Errorf(codes.Internal, "failed to check email: %v", err)
	}

	// 加密密码
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     userv1.UserRole_USER_ROLE_USER,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, "", status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return user.ToProto(), token, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (*userv1.User, string, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", status.Errorf(codes.NotFound, "user not found")
		}
		return nil, "", status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// 验证密码
	if !s.verifyPassword(password, user.Password) {
		return nil, "", status.Errorf(codes.Unauthenticated, "invalid password")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, "", status.Errorf(codes.PermissionDenied, "user account is disabled")
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, "", status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return user.ToProto(), token, nil
}

func (s *userService) GetUser(ctx context.Context, userID *uint64) (*userv1.User, error) {
	if userID == nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	var user models.User
	err := s.db.WithContext(ctx).First(&user, *userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return user.ToProto(), nil
}

func (s *userService) UpdateUser(ctx context.Context, id *uint64, updates map[string]interface{}) (*userv1.User, error) {
	operatorID, err := s.getCurrentUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get current user: %v", err)
	}

	// 确认被更新用户存在
	if _, err := s.GetUser(ctx, id); err != nil {
		return nil, err
	}

	// 判断操作者是否为超级管理员（依据网关注入的 x-user-role）
	isSuperAdmin := isUserRoleSuperAdmin(ctx)

	// 只有超级管理员可以更新敏感字段
	if _, ok := updates["role"]; ok && !isSuperAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can update user role")
	}
	if _, ok := updates["is_active"]; ok && !isSuperAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can update user status")
	}

	// 非超级管理员只能更新自己的基础信息（用户名/邮箱/头像），不能改别人
	if !isSuperAdmin && operatorID != *id {
		return nil, status.Errorf(codes.PermissionDenied, "only super admins can update other users")
	}

	var userModel models.User
	if err := s.db.WithContext(ctx).First(&userModel, *id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	if username, ok := updates["username"].(string); ok {
		userModel.Username = username
	}
	if email, ok := updates["email"].(string); ok {
		userModel.Email = email
	}
	if avatar, ok := updates["avatar"].(string); ok {
		userModel.Avatar = avatar
	}
	if role, ok := updates["role"].(userv1.UserRole); ok {
		userModel.Role = role
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		userModel.IsActive = isActive
	}

	if err := s.db.WithContext(ctx).Save(&userModel).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return s.GetUser(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context, page, pageSize int32) ([]*userv1.User, int32, int32, error) {
	offset := (page - 1) * pageSize

	var users []models.User
	var total int64

	if err := s.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to count users: %v", err)
	}

	if err := s.db.WithContext(ctx).Offset(int(offset)).Limit(int(pageSize)).Find(&users).Error; err != nil {
		return nil, 0, 0, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	userProtos := make([]*userv1.User, len(users))
	for i, user := range users {
		userProtos[i] = user.ToProto()
	}

	totalPages := int32(total) / pageSize
	if int32(total)%pageSize > 0 {
		totalPages++
	}

	return userProtos, int32(total), totalPages, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	if err := s.db.WithContext(ctx).Delete(&models.User{}, id).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return nil
}

func (s *userService) Auth(ctx context.Context, token string, tokenType userv1.TokenType, httpMethod, requestPath, clientIP, userAgent string) (*userv1.AuthResponse, error) {
	if token == "" {
		return &userv1.AuthResponse{
			Authenticated: false,
			Reason:        "missing token",
		}, nil
	}

	var user *userv1.User
	var scopes []userv1.Scope
	var expiresAt *timestamppb.Timestamp
	var err error

	if tokenType == userv1.TokenType_TOKEN_TYPE_UNSPECIFIED {
		// 根据令牌格式判断类型
		if strings.HasPrefix(token, "tk_") {
			// API令牌格式，直接使用API令牌验证
			user, scopes, expiresAt, err = s.validateAPIToken(ctx, token)
		} else {
			// 尝试JWT令牌验证
			user, expiresAt, err = s.validateUserToken(ctx, token)
			if err != nil {
				// JWT验证失败，尝试API令牌验证
				user, scopes, expiresAt, err = s.validateAPIToken(ctx, token)
			}
		}

		if err != nil {
			return &userv1.AuthResponse{
				Authenticated: false,
				Reason:        fmt.Sprintf("invalid token: %v", err),
			}, nil
		}
	} else {
		switch tokenType {
		case userv1.TokenType_TOKEN_TYPE_USER:
			user, expiresAt, err = s.validateUserToken(ctx, token)
		case userv1.TokenType_TOKEN_TYPE_API:
			user, scopes, expiresAt, err = s.validateAPIToken(ctx, token)
		default:
			return &userv1.AuthResponse{
				Authenticated: false,
				Reason:        "unsupported token type",
			}, nil
		}
	}

	if err != nil {
		return &userv1.AuthResponse{
			Authenticated: false,
			Reason:        fmt.Sprintf("invalid token: %v", err),
		}, nil
	}

	resp := &userv1.AuthResponse{
		Authenticated: true,
		User:          user,
		Scopes:        scopes,
		ExpiresAt:     expiresAt,
		Headers: map[string]string{
			"X-User-ID":   fmt.Sprintf("%d", user.Id),
			"X-Username":  user.Username,
			"X-User-Role": user.Role.String(),
		},
	}

	if len(scopes) > 0 {
		scopeStrs := make([]string, len(scopes))
		for i, scope := range scopes {
			scopeStrs[i] = scope.String()
		}
		resp.Headers["X-Scopes"] = strings.Join(scopeStrs, ",")
	}

	return resp, nil
}

func (s *userService) CreateAPIToken(ctx context.Context, name, description string, expiresIn userv1.TokenExpiration, scopes []userv1.Scope) (*userv1.CreateAPITokenResponse, error) {
	userID, err := s.getCurrentUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	now := time.Now()
	var expiresAt time.Time

	switch expiresIn {
	case userv1.TokenExpiration_TOKEN_EXPIRATION_NEVER:
		expiresAt = time.Time{}
	case userv1.TokenExpiration_TOKEN_EXPIRATION_ONE_DAY:
		expiresAt = now.Add(24 * time.Hour)
	case userv1.TokenExpiration_TOKEN_EXPIRATION_ONE_WEEK:
		expiresAt = now.Add(7 * 24 * time.Hour)
	case userv1.TokenExpiration_TOKEN_EXPIRATION_ONE_MONTH:
		expiresAt = now.Add(30 * 24 * time.Hour)
	case userv1.TokenExpiration_TOKEN_EXPIRATION_THREE_MONTHS:
		expiresAt = now.Add(90 * 24 * time.Hour)
	case userv1.TokenExpiration_TOKEN_EXPIRATION_ONE_YEAR:
		expiresAt = now.Add(365 * 24 * time.Hour)
	case userv1.TokenExpiration_TOKEN_EXPIRATION_UNSPECIFIED:
		fallthrough
	default:
		expiresAt = now.Add(24 * time.Hour)
	}

	token := s.generateRandomToken()
	tokenHash := s.hashToken(token)

	scopeStrs := make([]string, len(scopes))
	for i, scope := range scopes {
		scopeStrs[i] = scope.String()
	}

	apiToken := models.APIToken{
		UserID:      userID,
		Name:        name,
		TokenHash:   tokenHash,
		Description: description,
		Scopes:      pq.StringArray(scopeStrs),
	}

	if !expiresAt.IsZero() {
		apiToken.ExpiresAt = &expiresAt
	}

	if err := s.db.WithContext(ctx).Create(&apiToken).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create API token: %v", err)
	}

	return &userv1.CreateAPITokenResponse{
		Id:          apiToken.ID,
		Name:        name,
		Token:       token,
		ExpiresAt:   timestamppb.New(expiresAt),
		Description: description,
		Scopes:      scopes,
		CreatedAt:   timestamppb.New(now),
	}, nil
}

func (s *userService) ListAPITokens(ctx context.Context) ([]*userv1.CreateAPITokenResponse, error) {
	userID, err := s.getCurrentUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	var apiTokens []models.APIToken
	err = s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&apiTokens).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list API tokens: %v", err)
	}

	responses := make([]*userv1.CreateAPITokenResponse, len(apiTokens))
	for i, token := range apiTokens {
		scopes := make([]userv1.Scope, len(token.Scopes))
		for j, scope := range token.Scopes {
			if s, ok := userv1.Scope_value[scope]; ok {
				scopes[j] = userv1.Scope(s)
			}
		}

		responses[i] = &userv1.CreateAPITokenResponse{
			Id:          token.ID,
			Name:        token.Name,
			Token:       "",
			Description: token.Description,
			Scopes:      scopes,
			CreatedAt:   timestamppb.New(token.CreatedAt),
		}

		if token.ExpiresAt != nil {
			responses[i].ExpiresAt = timestamppb.New(*token.ExpiresAt)
		}
	}

	return responses, nil
}

func (s *userService) DeleteAPIToken(ctx context.Context, tokenID uint64) error {
	if err := s.db.WithContext(ctx).Delete(&models.APIToken{}, tokenID).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to delete API token: %v", err)
	}

	return nil
}

func (s *userService) generateJWT(userID uint64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) validateUserToken(ctx context.Context, token string) (*userv1.User, *timestamppb.Timestamp, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, nil, fmt.Errorf("invalid user_id in token")
	}

	userID := uint64(userIDFloat)

	var user models.User
	err = s.db.WithContext(ctx).First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	if exp, ok := claims["exp"].(float64); ok {
		expiresAt := time.Unix(int64(exp), 0)
		return user.ToProto(), timestamppb.New(expiresAt), nil
	}

	return user.ToProto(), nil, nil
}

func (s *userService) validateAPIToken(ctx context.Context, token string) (*userv1.User, []userv1.Scope, *timestamppb.Timestamp, error) {
	tokenHash := s.hashToken(token)

	var apiToken models.APIToken
	err := s.db.WithContext(ctx).Preload("User").Where("token_hash = ?", tokenHash).First(&apiToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil, fmt.Errorf("API token not found")
		}
		return nil, nil, nil, fmt.Errorf("failed to validate API token: %w", err)
	}

	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, nil, nil, fmt.Errorf("API token expired")
	}

	scopes := make([]userv1.Scope, len(apiToken.Scopes))
	for i, scope := range apiToken.Scopes {
		if s, ok := userv1.Scope_value[scope]; ok {
			scopes[i] = userv1.Scope(s)
		}
	}

	var expiresAt *timestamppb.Timestamp
	if apiToken.ExpiresAt != nil {
		expiresAt = timestamppb.New(*apiToken.ExpiresAt)
	}

	return apiToken.User.ToProto(), scopes, expiresAt, nil
}

func (s *userService) generateRandomToken() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			b[i] = letters[i%len(letters)]
			continue
		}
		b[i] = letters[n.Int64()]
	}

	return fmt.Sprintf("tk_%s", string(b))
}

func (s *userService) hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return fmt.Sprintf("sha256_%x", h.Sum(nil))
}

func (s *userService) getCurrentUserIDFromContext(ctx context.Context) (uint64, error) {
	if userID, ok := ctx.Value("user_id").(uint64); ok {
		return userID, nil
	}

	if token, ok := ctx.Value("token").(string); ok {
		user, _, err := s.validateUserToken(ctx, token)
		if err != nil {
			return 0, fmt.Errorf("invalid token: %w", err)
		}
		return user.Id, nil
	}

	return 0, fmt.Errorf("user not authenticated")
}

// hashPassword 使用 bcrypt 加密密码
func (s *userService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
