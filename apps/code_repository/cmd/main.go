package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
	"net/http"

	"xcoding/apps/code_repository/internal/config"
	"xcoding/apps/code_repository/internal/gateway"
	handler "xcoding/apps/code_repository/internal/grpc/handler"
	"xcoding/apps/code_repository/internal/grpc/interceptors"
	"xcoding/apps/code_repository/internal/models"
	"xcoding/apps/code_repository/internal/service"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
	projectv1 "xcoding/gen/go/project/v1"
	repodb "xcoding/pkg/db"
	"xcoding/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库（使用现有 db.NewGorm + AutoMigrate）
	dsn := cfg.Database.GetDSN()
	gormDB, err := repodb.NewGorm(dsn)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	if err := gormDB.AutoMigrate(&models.Repository{}, &models.RepositoryBranch{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Address, cfg.GRPC.Port)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port)

	grpcServer := startGRPCServer(cfg, gormDB.GetDB(), grpcAddr)
	httpServer := startHTTPServer(cfg, httpAddr)

	// 统一优雅关闭（包含数据库关闭）
	server.WaitForShutdown(
		grpcServer,
		httpServer,
		5*time.Second,
		func(ctx context.Context) error { return gormDB.Close() },
	)
}

func startGRPCServer(cfg *config.Config, gdb *gorm.DB, grpcAddr string) *grpc.Server {
	// 拨号 Project gRPC 客户端（用于权限校验/成员关系查询）
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

	// 初始化服务与 gRPC 处理器
	svc := service.NewCodeRepositoryService(gdb, projectClient)
	grpcServer := server.StartGRPCServer(
		grpcAddr,
		[]grpc.UnaryServerInterceptor{
			interceptors.RecoveryInterceptor(),
			interceptors.LoggingInterceptor(cfg),
			interceptors.MetricsInterceptor(cfg),
		},
		func(s *grpc.Server) {
			coderepositoryv1.RegisterCodeRepositoryServiceServer(s, handler.NewCodeRepositoryGRPCHandler(svc))
		},
	)
	return grpcServer
}

func startHTTPServer(cfg *config.Config, httpAddr string) *http.Server {
	// 初始化 HTTP 网关（grpc-gateway）
	mux := gateway.NewServeMux(cfg)
	dialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)
	if err := gateway.RegisterCodeRepositoryGateway(context.Background(), mux, dialAddr); err != nil {
		log.Fatalf("注册网关失败: %v", err)
	}
	return server.StartHTTPServerDefault(httpAddr, mux)
}
