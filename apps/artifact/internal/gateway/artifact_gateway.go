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

	"xcoding/apps/artifact/internal/config"
	artifactv1 "xcoding/gen/go/artifact/v1"
	"xcoding/pkg/gateway"
)

func RegisterArtifactGateway(ctx context.Context, mux *runtime.ServeMux, grpcAddr string) error {
	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}
	defer conn.Close()

	err = artifactv1.RegisterArtifactServiceHandler(ctx, mux, conn)
	if err != nil {
		return fmt.Errorf("failed to register artifact service handler: %w", err)
	}

	return nil
}

func NewServeMux(cfg *config.Config) *runtime.ServeMux {
	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	return runtime.NewServeMux(
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			return gateway.ForwardResponseHeaders(ctx, w, resp)
		}),
		runtime.WithIncomingHeaderMatcher(gateway.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithMetadata(gateway.MetadataFromRequest),
	)
}

func setResponseHeaders(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Echo user-related headers for tracing/debugging
	if sm, ok := runtime.ServerMetadataFromContext(ctx); ok {
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-id", "X-User-ID")
		setHeaderIfPresent(w, sm.HeaderMD, "x-username", "X-Username")
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-role", "X-User-Role")
		setHeaderIfPresent(w, sm.HeaderMD, "x-scopes", "X-Scopes")
	} else if md, ok := metadata.FromIncomingContext(ctx); ok {
		setHeaderIfPresent(w, md, "x-user-id", "X-User-ID")
		setHeaderIfPresent(w, md, "x-username", "X-Username")
		setHeaderIfPresent(w, md, "x-user-role", "X-User-Role")
		setHeaderIfPresent(w, md, "x-scopes", "X-Scopes")
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
	return gateway.CustomHeaderMatcher(key)
}

func customMetadataFunc(ctx context.Context, r *http.Request) metadata.MD {
	return gateway.MetadataFromRequest(ctx, r)
}
