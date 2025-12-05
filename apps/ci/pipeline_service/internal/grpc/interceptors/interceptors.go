package interceptors

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"xcoding/apps/ci/pipeline_service/internal/config"
)

// RecoveryInterceptor recovers from panics inside handlers
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered in %s: %v\n%s", info.FullMethod, r, string(debug.Stack()))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// LoggingInterceptor provides basic request logging in debug mode
func LoggingInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if cfg.Log.Level == "debug" {
			start := time.Now()
			resp, err = handler(ctx, req)
			log.Printf("method=%s duration=%s error=%v", info.FullMethod, time.Since(start), err)
			return resp, err
		}
		return handler(ctx, req)
	}
}

// MetricsInterceptor is a placeholder for metrics collection
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return handler(ctx, req)
	}
}
