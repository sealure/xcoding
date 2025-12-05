package server

import (
    "context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"
)

// StartGRPCServer starts a gRPC server on the given address with provided unary interceptors,
// registers services via the given register function, enables reflection and health service,
// and returns the server instance.
func StartGRPCServer(addr string, unaryInterceptors []grpc.UnaryServerInterceptor, register func(*grpc.Server)) *grpc.Server {
    var opts []grpc.ServerOption
    if len(unaryInterceptors) > 0 {
        opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
    }
    srv := grpc.NewServer(opts...)

    // Reflection for dev/debug
    reflection.Register(srv)

    // Standard gRPC health service set to SERVING
    hs := health.NewServer()
    healthgrpc.RegisterHealthServer(srv, hs)
    hs.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)

    // Service registration callback
    if register != nil {
        register(srv)
    }

    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("Failed to listen on %s: %v", addr, err)
    }
    log.Printf("Starting gRPC server on %s", addr)

    go func() {
        if err := srv.Serve(lis); err != nil {
            log.Fatalf("Failed to start gRPC server: %v", err)
        }
    }()

    return srv
}

// StartHTTPServerDefault starts an HTTP server with sensible defaults and returns it.
// It configures read/write/idle timeouts and starts the server in a goroutine.
func StartHTTPServerDefault(addr string, handler http.Handler) *http.Server {
    httpServer := &http.Server{
        Addr:         addr,
        Handler:      handler,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        log.Printf("Starting HTTP gateway server on %s", httpServer.Addr)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start HTTP gateway server: %v", err)
        }
    }()

    return httpServer
}

// ComputeLocalDialAddr returns a sane dial address for local gRPC connections.
// If listenAddr is blank/0.0.0.0/":", it uses 127.0.0.1:port; otherwise "listenAddr:port".
func ComputeLocalDialAddr(listenAddr string, port int) string {
    if listenAddr == "" || listenAddr == "0.0.0.0" || listenAddr == ":" {
        return net.JoinHostPort("127.0.0.1", intToString(port))
    }
    return net.JoinHostPort(listenAddr, intToString(port))
}

func intToString(i int) string {
    // avoid strconv import for minimal footprint
    // i is small (port), use simple conversion
    return fmtInt(i)
}

// minimal integer to string conversion without importing fmt/strconv in this file
// We keep this private and simple to avoid pulling extra deps; ports are small.
func fmtInt(i int) string {
    // Use a small buffer; ports are up to 5 digits
    var buf [6]byte
    pos := len(buf)
    if i == 0 {
        return "0"
    }
    for i > 0 {
        pos--
        buf[pos] = byte('0' + i%10)
        i /= 10
    }
    return string(buf[pos:])
}

// WaitForShutdown blocks until SIGINT/SIGTERM, then gracefully stops servers.
// It shuts down the HTTP server with the provided timeout, runs any extra closers,
// and finally performs gRPC GracefulStop.
func WaitForShutdown(grpcServer *grpc.Server, httpServer *http.Server, timeout time.Duration, extraClosers ...func(context.Context) error) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down servers...")

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    if httpServer != nil {
        if err := httpServer.Shutdown(ctx); err != nil {
            log.Printf("HTTP server forced to shutdown: %v", err)
        }
    }

    for _, closer := range extraClosers {
        if closer != nil {
            if err := closer(ctx); err != nil {
                log.Printf("Shutdown closer error: %v", err)
            }
        }
    }

    if grpcServer != nil {
        grpcServer.GracefulStop()
    }

    log.Println("Servers stopped")
}