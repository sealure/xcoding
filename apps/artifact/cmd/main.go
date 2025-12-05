package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"xcoding/apps/artifact/internal/config"
	"xcoding/apps/artifact/internal/gateway"
	handler "xcoding/apps/artifact/internal/grpc/handler"
	"xcoding/apps/artifact/internal/grpc/interceptors"
	"xcoding/apps/artifact/internal/models"
	"xcoding/apps/artifact/internal/service"
	artifactv1 "xcoding/gen/go/artifact/v1"
	projectv1 "xcoding/gen/go/project/v1"
	artdb "xcoding/pkg/db"
	"xcoding/pkg/server"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志级别
	switch cfg.Log.Level {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		log.SetFlags(log.LstdFlags)
	case "warn":
		log.SetOutput(os.Stdout)
	case "error":
		log.SetOutput(os.Stderr)
	}

	// 初始化数据库连接
	gormDB, err := artdb.NewGorm(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库模型
	if err := gormDB.AutoMigrate(
		&models.Registry{},
		&models.Namespace{},
		&models.Repository{},
		&models.Tag{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化 Project 客户端并注入服务层
	projectAddr := fmt.Sprintf("%s:%d", cfg.Project.Address, cfg.Project.Port)
	projectConn, err := grpc.DialContext(
		context.Background(),
		projectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("拨号 Project 服务失败: %v", err)
	}
	projectClient := projectv1.NewProjectServiceClient(projectConn)

	artifactService := service.NewArtifactService(gormDB.GetDB(), projectClient)

	// 计算地址
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Address, cfg.GRPC.Port)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port)

	// 启动 gRPC 服务器（统一抽象）
	grpcServer := server.StartGRPCServer(
		grpcAddr,
		[]grpc.UnaryServerInterceptor{
			interceptors.RecoveryInterceptor(),
			interceptors.LoggingInterceptor(cfg),
			interceptors.MetricsInterceptor(),
			interceptors.NewAuthInterceptor(artifactService, cfg).UnaryServerInterceptor(),
		},
		func(s *grpc.Server) {
			artifactv1.RegisterArtifactServiceServer(s, handler.NewArtifactGRPCHandler(artifactService))
		},
	)

	// 启动 HTTP 网关（复用网关+通用HTTP启动）
	mux := gateway.NewServeMux(cfg)
	dialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)
	conn, err := grpc.DialContext(context.Background(), dialAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server (%s): %v", dialAddr, err)
	}
	if err := artifactv1.RegisterArtifactServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("Failed to register artifact service handler: %v", err)
	}
	httpServer := server.StartHTTPServerDefault(httpAddr, mux)

	// 统一优雅关闭（包含数据库关闭）
	server.WaitForShutdown(
		grpcServer,
		httpServer,
		30*time.Second,
		func(ctx context.Context) error { return gormDB.Close() },
	)
}
