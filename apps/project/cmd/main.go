package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"xcoding/apps/project/internal/config"
	"xcoding/apps/project/internal/gateway"
	handler "xcoding/apps/project/internal/grpc/handler"
	"xcoding/apps/project/internal/grpc/interceptors"
	"xcoding/apps/project/internal/models"
	"xcoding/apps/project/internal/service"
	projectv1 "xcoding/gen/go/project/v1"
	projdb "xcoding/pkg/db"
	"xcoding/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	dsn := cfg.Database.GetDSN()
	gormDB, err := projdb.NewGorm(dsn)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	if err := gormDB.AutoMigrate(&models.Project{}, &models.ProjectMember{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	projectService := service.NewProjectService(gormDB.GetDB())

	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Address, cfg.GRPC.Port)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port)

	// 启动 gRPC 服务器（抽象库）
	grpcServer := server.StartGRPCServer(
		grpcAddr,
		[]grpc.UnaryServerInterceptor{
			interceptors.RecoveryInterceptor(),
			interceptors.LoggingInterceptor(cfg),
			interceptors.MetricsInterceptor(cfg),
		},
		func(s *grpc.Server) {
			projectv1.RegisterProjectServiceServer(s, handler.NewProjectGRPCHandler(projectService))
		},
	)

	// 启动 HTTP 网关（现有健康检查挂载保持不变）
	mux := gateway.NewServeMux(cfg)
	dialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)
	conn, err := grpc.DialContext(context.Background(), dialAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接gRPC服务器失败: %v", err)
	}
	if err := projectv1.RegisterProjectServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("注册项目服务处理器失败: %v", err)
	}
	healthServer := gateway.NewHealthServer(cfg)
	healthServer.RegisterHandlers(mux)
	httpServer := server.StartHTTPServerDefault(httpAddr, mux)
	// 设置健康检查为就绪
	healthServer.SetReady(true)

	// 统一优雅关闭（包含数据库关闭）
	server.WaitForShutdown(
		grpcServer,
		httpServer,
		5*time.Second,
		func(ctx context.Context) error { return gormDB.Close() },
	)
}
