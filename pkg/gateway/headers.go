package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// CustomHeaderMatcher forwards selected headers to gRPC metadata.
func CustomHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization",
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Request-ID",
		"User-Agent",
		"X-User-ID",
		"X-User-Role",
		"X-Username",
		"X-Scopes",
		"X-Project-ID":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// MetadataFromRequest builds gRPC metadata from incoming HTTP request.
func MetadataFromRequest(ctx context.Context, r *http.Request) metadata.MD {
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
	projectID := r.Header.Get("X-Project-ID")

	// Prefer USER_ROLE_SUPER_ADMIN among duplicated role headers.
	userRole := ""
	for _, v := range r.Header.Values("X-User-Role") {
		vv := strings.TrimSpace(v)
		if vv == "" {
			continue
		}
		if vv == "USER_ROLE_SUPER_ADMIN" || vv == "SUPER_ADMIN" {
			userRole = "USER_ROLE_SUPER_ADMIN"
			break
		}
		if userRole == "" {
			userRole = vv
		}
	}
	if userRole == "" {
		userRole = r.Header.Get("X-User-Role")
	}

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
	if projectID != "" {
		md.Set("x-project-id", projectID)
	}

	// Prefer original URI/method from proxy if present
	origURI := r.Header.Get("X-Original-URI")
	origMethod := r.Header.Get("X-Original-Method")
	requestPath := r.URL.Path
	httpMethod := r.Method
	if origURI != "" { requestPath = origURI }
	if origMethod != "" { httpMethod = origMethod }
	md.Set("x-request-path", requestPath)
	md.Set("x-request-method", httpMethod)
	return md
}

// ForwardResponseHeaders writes security headers and echoes select metadata into HTTP response.
func ForwardResponseHeaders(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	if sm, ok := runtime.ServerMetadataFromContext(ctx); ok {
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-id", "X-User-ID")
		setHeaderIfPresent(w, sm.HeaderMD, "x-username", "X-Username")
		setHeaderIfPresent(w, sm.HeaderMD, "x-user-role", "X-User-Role")
		setHeaderIfPresent(w, sm.HeaderMD, "x-scopes", "X-Scopes")
		setHeaderIfPresent(w, sm.HeaderMD, "x-project-id", "X-Project-ID")
	} else if md, ok := metadata.FromIncomingContext(ctx); ok {
		setHeaderIfPresent(w, md, "x-user-id", "X-User-ID")
		setHeaderIfPresent(w, md, "x-username", "X-Username")
		setHeaderIfPresent(w, md, "x-user-role", "X-User-Role")
		setHeaderIfPresent(w, md, "x-scopes", "X-Scopes")
		setHeaderIfPresent(w, md, "x-project-id", "X-Project-ID")
	}
	return nil
}

func setHeaderIfPresent(w http.ResponseWriter, md metadata.MD, key string, header string) {
	vals := md.Get(key)
	if len(vals) > 0 {
		w.Header().Set(header, vals[0])
	}
}