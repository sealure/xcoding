package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xcoding/apps/code_repository/internal/config"
)

// RecoveryInterceptor 捕获panic
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

// LoggingInterceptor 记录RPC调用信息
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

// MetricsInterceptor 预留
func MetricsInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}
