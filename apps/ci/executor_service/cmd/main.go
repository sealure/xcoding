package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"xcoding/apps/ci/executor_service/internal/config"
	"xcoding/apps/ci/executor_service/internal/consumer"
	"xcoding/apps/ci/executor_service/internal/gateway"
	"xcoding/apps/ci/executor_service/internal/service"
	"xcoding/apps/ci/executor_service/internal/ws"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"
	cddb "xcoding/pkg/db"
	"xcoding/pkg/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type executorServer struct {
	civ1.UnimplementedExecutorServiceServer
}

var execSvc *service.ExecutorService

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load executor config: %v", err)
	}

	gormDB, err := cddb.NewGorm(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect executor database: %v", err)
	}

	if err := gormDB.AutoMigrate(
		&models.Build{}, &models.BuildSnapshot{}, &models.BuildJob{}, &models.BuildJobEdge{}, &models.BuildStep{}, &models.BuildStepLogChunk{},
	); err != nil {
		log.Fatalf("Executor migrate failed: %v", err)
	}

	grpcAddr := cfg.GRPCAddr()
	httpAddr := cfg.HTTPAddr()

	grpcServer := server.StartGRPCServer(grpcAddr, nil, func(s *grpc.Server) {
		execSvc = service.New(gormDB.GetDB())
		civ1.RegisterExecutorServiceServer(s, execSvc)
	})

	mux := gateway.NewServeMux(cfg)
	dialAddr := server.ComputeLocalDialAddr(cfg.GRPC.Address, cfg.GRPC.Port)
	conn, err := grpc.DialContext(context.Background(), dialAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("executor: dial grpc %s: %v", dialAddr, err)
	}
	if err := civ1.RegisterExecutorServiceHandler(context.Background(), mux, conn); err != nil {
		log.Fatalf("executor: register gateway: %v", err)
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/ci_service/api/v1/executor/ws/builds/", ws.NewHandler(gormDB.GetDB()))
	rootMux.Handle("/", mux)

	httpServer := server.StartHTTPServerDefault(httpAddr, rootMux)

	execClient := civ1.NewExecutorServiceClient(conn)
	url := strings.TrimSpace(cfg.Queue.URL)
	if url == "" {
		log.Printf("executor: RabbitMQ URL not set; queue consumer disabled")
		server.WaitForShutdown(grpcServer, httpServer, cfg.ShutdownTimeout())
		return
	}
	qname := strings.TrimSpace(cfg.Queue.Queue)
	if qname == "" {
		qname = "ci_builds"
	}
	qc := consumer.NewQueueConsumer(url, qname, execClient, gormDB.GetDB(), os.Getenv("POD_NAMESPACE"))
	if err := qc.Start(context.Background()); err != nil {
		log.Printf("executor: queue start error: %v", err)
	}
	server.WaitForShutdown(grpcServer, httpServer, cfg.ShutdownTimeout(), func(ctx context.Context) error { qc.Close(); return nil })
}
