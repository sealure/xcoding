package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"xcoding/apps/code_repository/internal/config"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
	"xcoding/pkg/gateway"
)

// RegisterCodeRepositoryGateway 将 gRPC 服务通过 grpc-gateway 暴露为 HTTP 接口
// - 仅负责注册路由与转发，不做业务参数校验（校验在 Handler/Service 层）
// - 透传上游请求头（如认证/项目上下文）到下游 gRPC 服务
func RegisterCodeRepositoryGateway(ctx context.Context, mux *runtime.ServeMux, grpcAddr string) error {
	// 建立到 gRPC 服务的连接（insecure，仅用于本地或受控环境）
	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}
	// 保持连接在网关生命周期内，不在此处关闭

	// 注册 code_repository 的 HTTP 路由到 ServeMux
	// 生成的 handler 会将 HTTP 请求转换为对应的 gRPC 方法调用
	if err := coderepositoryv1.RegisterCodeRepositoryServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("failed to register code_repository service handler: %w", err)
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
			return gateway.ForwardResponseHeaders(ctx, w, resp)
		}),
		runtime.WithIncomingHeaderMatcher(gateway.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithMetadata(gateway.MetadataFromRequest),
	)
}
