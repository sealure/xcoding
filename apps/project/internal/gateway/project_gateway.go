package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"xcoding/apps/project/internal/config"
	projectv1 "xcoding/gen/go/project/v1"
)

func RegisterProjectGateway(ctx context.Context, mux *runtime.ServeMux, grpcAddr string) error {
	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}
	defer conn.Close()

	if err := projectv1.RegisterProjectServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("failed to register project service handler: %w", err)
	}
	return nil
}

func NewServeMux(cfg *config.Config) *runtime.ServeMux {
	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{DiscardUnknown: true},
	}

	return runtime.NewServeMux(
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			return setResponseHeaders(ctx, w, resp)
		}),
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithMetadata(customMetadataFunc),
	)
}

func setResponseHeaders(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Echo user-related headers back to client for tracing/testing
	if sm, ok := runtime.ServerMetadataFromContext(ctx); ok {
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-id", "X-User-ID")
		setHeaderIfPresent(w, sm.HeaderMD, "x-username", "X-Username")
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-role", "X-User-Role")
		setHeaderIfPresent(w, sm.HeaderMD, "x-scopes", "X-Scopes")
	}
	return nil
}

func setHeaderIfPresent(w http.ResponseWriter, md metadata.MD, key string, header string) {
	vals := md.Get(key)
	if len(vals) > 0 {
		w.Header().Set(header, vals[0])
	}
}

func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization",
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Request-ID",
		"User-Agent",
		"X-User-ID",
		"X-User-Role",
		"X-Username",
		"X-Scopes":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func customMetadataFunc(ctx context.Context, r *http.Request) metadata.MD {
	md := metadata.MD{}

	if auth := r.Header.Get("Authorization"); auth != "" {
		md.Set("authorization", auth)
	}

	if clientIP := r.Header.Get("X-Real-IP"); clientIP != "" {
		md.Set("x-client-ip", clientIP)
	} else if clientIP = r.Header.Get("X-Forwarded-For"); clientIP != "" {
		md.Set("x-client-ip", clientIP)
	} else {
		md.Set("x-client-ip", r.RemoteAddr)
	}

	if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
		md.Set("x-user-agent", userAgent)
	}

	userID := r.Header.Get("X-User-ID")
	username := r.Header.Get("X-Username")
	userRole := r.Header.Get("X-User-Role")
	scopes := r.Header.Get("X-Scopes")
	if userID != "" {
		md.Set("x-user-id", userID)
	}
	if username != "" {
		md.Set("x-username", username)
	}
	if userRole != "" {
		md.Set("x-user-role", userRole)
	}
	if scopes != "" {
		md.Set("x-scopes", scopes)
	}

	md.Set("x-request-path", r.URL.Path)
	md.Set("x-request-method", r.Method)

	return md
}
