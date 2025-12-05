package gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"xcoding/apps/ci/pipeline_service/internal/config"
	civ1 "xcoding/gen/go/ci/v1"
	xgw "xcoding/pkg/gateway"
)

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
			return xgw.ForwardResponseHeaders(ctx, w, resp)
		}),
		runtime.WithIncomingHeaderMatcher(xgw.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithMetadata(xgw.MetadataFromRequest),
	)
}

// Helper to register handlers if needed elsewhere
func Register(ctx context.Context, mux *runtime.ServeMux, conn civ1.PipelineServiceClient) error {
	// Not used by main.go; kept for symmetry if needed
	_ = conn
	return nil
}
