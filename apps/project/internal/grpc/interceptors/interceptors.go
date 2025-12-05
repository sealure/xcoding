package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xcoding/apps/project/internal/config"
)

// RecoveryInterceptor 捕获panic，避免服务崩溃
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered in gRPC call %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// LoggingInterceptor 记录RPC调用信息和耗时（在debug下更详细）
func LoggingInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.Log.Level == "debug" {
			log.Printf("gRPC call: %s", info.FullMethod)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if cfg.Log.Level == "debug" {
			log.Printf("gRPC done: %s, duration: %v, err: %v", info.FullMethod, duration, err)
		}

		return resp, err
	}
}

// MetricsInterceptor 预留用于指标收集
func MetricsInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: 集成Prometheus或OpenTelemetry指标
		return handler(ctx, req)
	}
}
