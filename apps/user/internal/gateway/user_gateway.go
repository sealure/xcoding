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

	"xcoding/apps/user/internal/config"
	userv1 "xcoding/gen/go/user/v1"
)

func RegisterUserGateway(ctx context.Context, mux *runtime.ServeMux, grpcAddr string) error {
	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server: %w", err)
	}
	defer conn.Close()

	err = userv1.RegisterUserServiceHandler(ctx, mux, conn)
	if err != nil {
		return fmt.Errorf("failed to register user service handler: %w", err)
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

	// 若返回值为 AuthResponse，则把其中的 Headers 写入 HTTP 响应头，供 APISIX forward-auth 使用
	if ar, ok := resp.(*userv1.AuthResponse); ok {
		for k, v := range ar.Headers {
			if v != "" {
				w.Header().Set(k, v)
			}
		}
		if ar.User != nil {
			if w.Header().Get("X-User-ID") == "" {
				w.Header().Set("X-User-ID", fmt.Sprintf("%d", ar.User.Id))
			}
			if w.Header().Get("X-Username") == "" {
				w.Header().Set("X-Username", ar.User.Username)
			}

			if w.Header().Get("X-User-Role") == "" {
				w.Header().Set("X-User-Role", ar.User.Role.String())
			}
		}
	}

	return nil
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

	// 提取Authorization头中的token
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

	// 将 forward-auth 生成的用户头注入到 metadata，便于调试/后续处理
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

	// 打印收到的用户头，便于验证 APISIX forward-auth 是否生效
	if userID != "" || username != "" || userRole != "" || scopes != "" {
		fmt.Printf("ForwardAuth headers: user_id=%s username=%s role=%s scopes=%s\n", userID, username, userRole, scopes)
	}

	// 记录请求的路径和方法
	md.Set("x-request-path", r.URL.Path)
	md.Set("x-request-method", r.Method)

	return md
}
