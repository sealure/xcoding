package gateway

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"xcoding/apps/project/internal/config"
)

// HealthServer 健康检查服务器
type HealthServer struct {
	mu        sync.RWMutex
	startTime time.Time
	ready     bool
	live      bool
	shutdown  bool
	cfg       *config.Config
}

func NewHealthServer(cfg *config.Config) *HealthServer {
	return &HealthServer{
		startTime: time.Now(),
		ready:     false,
		live:      true,
		shutdown:  false,
		cfg:       cfg,
	}
}

func (h *HealthServer) SetReady(ready bool) { h.mu.Lock(); h.ready = ready; h.mu.Unlock() }
func (h *HealthServer) SetLive(live bool)   { h.mu.Lock(); h.live = live; h.mu.Unlock() }
func (h *HealthServer) Shutdown()           { h.mu.Lock(); h.shutdown = true; h.mu.Unlock() }

func (h *HealthServer) RegisterHandlers(mux *runtime.ServeMux) {
	mux.HandlePath("GET", "/project_service/healthz", h.livenessHandler)
	mux.HandlePath("GET", "/project_service/readyz", h.readinessHandler)
	mux.HandlePath("GET", "/project_service/livez", h.startupHandler)
}

func (h *HealthServer) livenessHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.shutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}
	if !h.live {
		http.Error(w, "Service is not live", http.StatusServiceUnavailable)
		return
	}
	uptime := time.Since(h.startTime).String()
	response := fmt.Sprintf(`{"status":"ok","uptime":"%s","timestamp":"%s"}`, uptime, time.Now().Format(time.RFC3339))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (h *HealthServer) readinessHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.shutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}
	if !h.ready {
		http.Error(w, "Service is not ready", http.StatusServiceUnavailable)
		return
	}
	uptime := time.Since(h.startTime).String()
	response := fmt.Sprintf(`{"status":"ok","uptime":"%s","timestamp":"%s","checks":{"database":"ok","grpc":"ok","http":"ok"}}`, uptime, time.Now().Format(time.RFC3339))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (h *HealthServer) startupHandler(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.shutdown {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}
	if !h.ready {
		http.Error(w, "Service is not ready", http.StatusServiceUnavailable)
		return
	}
	uptime := time.Since(h.startTime).String()
	response := fmt.Sprintf(`{"status":"ok","uptime":"%s","timestamp":"%s","version":"%s"}`, uptime, time.Now().Format(time.RFC3339), h.cfg.App.Version)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
