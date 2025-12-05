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

	"xcoding/apps/artifact/internal/config"
	"xcoding/apps/artifact/internal/service"
)

/*
AuthInterceptor（认证拦截器）：

- 在调用需要认证的 RPC 方法前验证令牌
- 从请求头中提取 Authorization 令牌
- 调用 userService.Auth() 方法验证令牌有效性
- 将用户信息添加到上下文中，供后续处理使用
*/
type AuthInterceptor struct {
	artifactService service.ArtifactService
	publicMethods   map[string]bool
	cfg             *config.Config
}

// NewAuthInterceptor 创建认证拦截器
func NewAuthInterceptor(artifactService service.ArtifactService, cfg *config.Config) *AuthInterceptor {
	publicMethods := map[string]bool{
		// "/artifact.v1.ArtifactService/ListRegistries":   true,
		// "/artifact.v1.ArtifactService/ListNamespaces":   true,
		// "/artifact.v1.ArtifactService/ListRepositories": true,
		// "/artifact.v1.ArtifactService/ListTags":         true,
		// "/artifact.v1.ArtifactService/GetImageManifest": true,
		// "/artifact.v1.ArtifactService/GetImageLayer":    true,
		"/grpc.health.v1.Health/Check": true,
		"/grpc.health.v1.Health/Watch": true,
	}

	return &AuthInterceptor{
		artifactService: artifactService,
		publicMethods:   publicMethods,
		cfg:             cfg,
	}
}

/*
UnaryServerInterceptor（一元服务器拦截器）：
一元 RPC 是指最简单的请求-响应模式，客户端发送一个请求，服务器返回一个响应。
gRPC 中，当客户端发起一个一元 RPC 调用时，请求会经过一系列的拦截器处理，然后才到达实际的服务处理方法。拦截器链的执行顺序如下：
1.客户端请求 → 拦截器1 → 拦截器2 → ... → 拦截器N → 实际服务方法
2.响应返回时则按相反顺序：实际服务方法 → 拦截器N → ... → 拦截器2 → 拦截器1 → 客户端

gRPC 支持四种不同的调用类型：
1. 一元（Unary）调用 ：
  - 客户端发送一个请求，服务器返回一个响应
  - 这是最简单的调用模式
  - 例如： GetRegistry(GetRegistryRequest) returns (GetRegistryResponse)

2.服务器流式（Server Streaming）调用 ：
  - 客户端发送一个请求，服务器返回一个流（多个响应）
  - 例如： ListRegistries(ListRegistriesRequest) returns (stream ListRegistriesResponse)

3.客户端流式（Client Streaming）调用 ：
  - 客户端发送一个流（多个请求），服务器返回一个响应
  - 例如： CreateRegistries(stream CreateRegistryRequest) returns (CreateRegistriesResponse)

4.双向流式（Bidirectional Streaming）调用 ：
  - 客户端和服务器都可以独立地向对方发送流
  - 例如： Chat(stream ChatRequest) returns (stream ChatResponse)
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

		// TODO: 实际项目中应该调用用户服务进行认证
		// 这里简化处理，假设令牌有效
		// 可以使用获取的请求信息进行日志记录或审计
		_ = httpMethod
		_ = requestPath
		_ = clientIP
		_ = userAgent

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

- 记录 RPC 请求的调用信息和执行时间
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
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: 实现指标收集
		return handler(ctx, req)
	})
}
