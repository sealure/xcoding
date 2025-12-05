package interceptors

import (
	"context"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"xcoding/apps/user/internal/config"
	"xcoding/apps/user/internal/service"
	userv1 "xcoding/gen/go/user/v1"
)

/*
AuthInterceptor（认证拦截器）：

- 在调用需要认证的 RPC 方法前验证令牌
- 从请求头中提取 Authorization 令牌
- 调用 userService.Auth() 方法验证令牌有效性
- 将用户信息添加到上下文中，供后续处理使用
*/
type AuthInterceptor struct {
	userService   service.UserService
	publicMethods map[string]bool
	cfg           *config.Config
}

// NewAuthInterceptor 创建认证拦截器
func NewAuthInterceptor(userService service.UserService, cfg *config.Config) *AuthInterceptor {
	publicMethods := map[string]bool{
		"/user.v1.UserService/Register": true,
		"/user.v1.UserService/Login":    true,
		"/user.v1.UserService/Auth":     true,
		"/grpc.health.v1.Health/Check":  true,
		"/grpc.health.v1.Health/Watch":  true,
	}

	return &AuthInterceptor{
		userService:   userService,
		publicMethods: publicMethods,
		cfg:           cfg,
	}
}

/*
UnaryServerInterceptor（一元服务器拦截器）：
一元 RPC 是最简单的请求-响应模式：客户端发送一个请求，服务器返回一个响应。
在 gRPC 中，客户端发起一元 RPC 调用时，请求会经过一系列拦截器，再到达实际服务方法。执行顺序如下：
1. 客户端请求 → 拦截器1 → 拦截器2 → ... → 拦截器N → 实际服务方法
2. 响应返回时则反向经过：实际服务方法 → 拦截器N → ... → 拦截器2 → 拦截器1 → 客户端

gRPC 支持四种调用类型：
1. 一元（Unary）：客户端发一个请求，服务器回一个响应。例如：GetUser(GetUserRequest) returns (GetUserResponse)
2. 服务器流式（Server Streaming）：客户端发一个请求，服务器返回一个响应流。例如：ListUsers(ListUsersRequest) returns (stream ListUsersResponse)
3. 客户端流式（Client Streaming）：客户端发送一个请求流，服务器返回一个响应。例如：CreateUsers(stream CreateUserRequest) returns (CreateUsersResponse)
4. 双向流式（Bidirectional Streaming）：客户端和服务器都可独立地向对方发送流。例如：Chat(stream ChatRequest) returns (stream ChatResponse)
*/
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 检查是否是公开方法
		if a.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 从元数据中获取令牌
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// 获取Authorization头
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		// 提取令牌
		token := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if token == authHeaders[0] {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		// 获取请求信息
		httpMethod := ""
		requestPath := ""
		clientIP := ""
		userAgent := ""

		if methodHeaders := md.Get("x-request-method"); len(methodHeaders) > 0 {
			httpMethod = methodHeaders[0]
		}
		if pathHeaders := md.Get("x-request-path"); len(pathHeaders) > 0 {
			requestPath = pathHeaders[0]
		}
		if ipHeaders := md.Get("x-client-ip"); len(ipHeaders) > 0 {
			clientIP = ipHeaders[0]
		}
		if uaHeaders := md.Get("x-user-agent"); len(uaHeaders) > 0 {
			userAgent = uaHeaders[0]
		}

		// 验证令牌类型
		if token == "" {
			return nil, status.Errorf(codes.Unauthenticated, "无效的令牌")
		}

		// 调用认证服务，使用UNSPECIFIED类型让服务自动检测令牌类型
		authResp, err := a.userService.Auth(ctx, token, userv1.TokenType_TOKEN_TYPE_UNSPECIFIED, httpMethod, requestPath, clientIP, userAgent)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// 检查认证结果
		if !authResp.Authenticated {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %s", authResp.Reason)
		}

		// 将用户信息添加到上下文
		if authResp.User != nil {
			ctx = context.WithValue(ctx, "user", authResp.User)
			ctx = context.WithValue(ctx, "user_id", authResp.User.Id)
		}

		// 将权限范围添加到上下文
		if len(authResp.Scopes) > 0 {
			ctx = context.WithValue(ctx, "scopes", authResp.Scopes)
		}

		// 调用处理程序
		return handler(ctx, req)
	}
}

/*
RecoveryInterceptor（恢复拦截器）：

- 捕获 RPC 处理过程中的 panic 异常
- 记录异常信息
- 返回内部服务器错误状态给客户端
*/
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	})
}

/*
LoggingInterceptor（日志拦截器）：

- 记录 RPC 请求的调用信息与执行时间
- 仅在 debug 模式下输出详细日志
*/
func LoggingInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.Log.Level == "debug" {
			log.Printf("gRPC call: %s", info.FullMethod)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if cfg.Log.Level == "debug" {
			log.Printf("gRPC call completed: %s, duration: %v, error: %v", info.FullMethod, duration, err)
		}

		return resp, err
	}
}

/*
MetricsInterceptor（指标拦截器）：

- 预留用于收集指标数据的接口（目前为空实现）
*/
func MetricsInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 在实际应用中，这里可以添加指标收集逻辑
		return handler(ctx, req)
	}
}
