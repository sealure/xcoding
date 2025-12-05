package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"xcoding/apps/ci/pipeline_service/internal/config"
	"xcoding/apps/ci/pipeline_service/internal/gateway"
	handler "xcoding/apps/ci/pipeline_service/internal/grpc/handler"
	"xcoding/apps/ci/pipeline_service/internal/grpc/interceptors"
	"xcoding/apps/ci/pipeline_service/internal/models"
	"xcoding/apps/ci/pipeline_service/internal/service"
	civ1 "xcoding/gen/go/ci/v1"
	civ1exec "xcoding/gen/go/ci/v1"
	projectv1 "xcoding/gen/go/project/v1"
	cddb "xcoding/pkg/db"
	"xcoding/pkg/server"
)

func main() {
	// 加载配置
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
	gormDB, err := cddb.NewGorm(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移模型（占位：Pipeline/Build/Schedule）
	if err := gormDB.AutoMigrate(
		&models.Pipeline{},
		&models.PipelineSchedule{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 拨号 Project 服务（用于权限与成员校验）
	projectAddr := fmt.Sprintf("%s:%d", cfg.Project.Address, cfg.Project.Port)
	projectConn, err := grpc.DialContext(
		context.Background(),
		projectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial Project service: %v", err)
	}
	projectClient := projectv1.NewProjectServiceClient(projectConn)

	// 拨号 Executor 服务（用于创建构建与运行态数据）
	executorAddr := fmt.Sprintf("%s:%d", cfg.Executor.Address, cfg.Executor.Port)
	executorConn, err := grpc.DialContext(
		context.Background(),
		executorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial Executor service: %v", err)
	}
	executorClient := civ1exec.NewExecutorServiceClient(executorConn)

	// 初始化服务
	svc := service.NewPipelineService(gormDB.GetDB(), projectClient, executorClient)

	var rmqClose func(context.Context) error
	{
		url := cfg.Queue.URL
		enabled := cfg.Queue.Enabled || url != ""
		if enabled {
			qname := cfg.Queue.Queue
			if qname == "" {
				qname = "ci_builds"
			}
			q, err := service.NewRabbitMQBuildQueue(url, qname)
			if err != nil {
				log.Fatalf("初始化 RabbitMQ 失败: %v", err)
			}
			be := handler.NewBuildExecutorHandler(q)
			if err := be.Init(context.Background()); err != nil {
				log.Fatalf("注册队列失败: %v", err)
			}
			rmqClose = func(ctx context.Context) error { q.Close(); return nil }
			log.Printf("RabbitMQ 队列已启用: url=%s queue=%s", url, qname)
		} else {
			log.Printf("RabbitMQ 队列未启用")
		}
	}

	// 计算地址
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Address, cfg.GRPC.Port)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Address, cfg.HTTP.Port)

	// 启动 gRPC 服务器
	grpcServer := server.StartGRPCServer(
		grpcAddr,
		[]grpc.UnaryServerInterceptor{
			interceptors.RecoveryInterceptor(),
			interceptors.LoggingInterceptor(cfg),
			interceptors.MetricsInterceptor(),
		},
		func(s *grpc.Server) {
			civ1.RegisterPipelineServiceServer(s, handler.NewPipelineGRPCHandler(svc))
		},
	)

	// 启动 HTTP 网关
	mux := gateway.NewServeMux(cfg)
	dialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)
	conn, err := grpc.DialContext(context.Background(), dialAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server (%s): %v", dialAddr, err)
	}
	if err := civ1.RegisterPipelineServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("Failed to register pipeline service handler: %v", err)
	}
	httpServer := server.StartHTTPServerDefault(httpAddr, mux)

	// 优雅关闭
	// 优雅关闭：数据库 + 队列
	if rmqClose != nil {
		server.WaitForShutdown(
			grpcServer,
			httpServer,
			30*time.Second,
			func(ctx context.Context) error { return gormDB.Close() },
			rmqClose,
		)
	} else {
		server.WaitForShutdown(
			grpcServer,
			httpServer,
			30*time.Second,
			func(ctx context.Context) error { return gormDB.Close() },
		)
	}
}
