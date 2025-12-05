package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"net/http"
	"xcoding/apps/user/internal/config"
	"xcoding/apps/user/internal/gateway"
	handler "xcoding/apps/user/internal/grpc/handler"
	"xcoding/apps/user/internal/grpc/interceptors"
	"xcoding/apps/user/internal/models"
	"xcoding/apps/user/internal/service"
	userv1 "xcoding/gen/go/user/v1"
	userdb "xcoding/pkg/db"
	"xcoding/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	dsn := cfg.Database.GetDSN()
	gormDB, err := userdb.NewGorm(dsn)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	// 同步迁移用户与 API 令牌表结构，避免创建令牌时报错
	if err := gormDB.AutoMigrate(&models.User{}, &models.APIToken{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userService := service.NewUserService(gormDB.GetDB(), cfg.JWT.Secret)
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Address, cfg.GRPC.Port)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port)

	// gRPC 启动（抽象库）
	grpcServer := server.StartGRPCServer(
		grpcAddr,
		[]grpc.UnaryServerInterceptor{
			interceptors.RecoveryInterceptor(),
			interceptors.LoggingInterceptor(cfg),
			interceptors.MetricsInterceptor(cfg),
			interceptors.NewAuthInterceptor(userService, cfg).UnaryServerInterceptor(),
		},
		func(s *grpc.Server) { userv1.RegisterUserServiceServer(s, handler.NewUserGRPCHandler(userService)) },
	)

	// HTTP 网关启动（抽象库）
	httpServer := startHTTPGatewayServer(cfg, userService, httpAddr)

	// 统一优雅关闭（包含数据库关闭）
	server.WaitForShutdown(
		grpcServer,
		httpServer,
		5*time.Second,
		func(ctx context.Context) error { return gormDB.Close() },
	)
}

// startGRPCServer 启动gRPC服务器
// startHTTPGatewayServer 启动HTTP网关服务器
func startHTTPGatewayServer(cfg *config.Config, userService service.UserService, httpAddr string) *http.Server {
	// 计算用于网关拨号的 gRPC 目标地址
	grpcDialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)

	// 创建gRPC连接 - 不要在这里关闭连接，因为HTTP网关需要持续使用它
	conn, err := grpc.DialContext(
		context.Background(),
		grpcDialAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial gRPC server (%s): %v", grpcDialAddr, err)
	}

	// 创建gRPC-Gateway的ServeMux，已集成中间件
	mux := gateway.NewServeMux(cfg)

	// 注册gRPC服务到网关
	if err := userv1.RegisterUserServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("Failed to register user service handler: %v", err)
	}

	// 启动HTTP网关服务器（抽象库）
	return server.StartHTTPServerDefault(httpAddr, mux)
}
