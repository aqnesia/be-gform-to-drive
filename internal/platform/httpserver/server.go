package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"automate-gform-drive/internal/health"
	"automate-gform-drive/internal/job"
	"automate-gform-drive/internal/webhook"
)

type Server struct {
	server *http.Server
}

func New(port string, webhookSecret string, repo *job.Repository, worker *job.Worker) *Server {
	mux := http.NewServeMux()

	healthHandler := health.NewHandler()
	webhookHandler := webhook.NewHandler(webhookSecret, repo, worker)

	mux.Handle("GET /", healthHandler)
	mux.Handle("GET /health", healthHandler)
	mux.Handle("POST /api/webhook/gform", webhookHandler)

	// Logging middleware
	handler := loggingMiddleware(mux)

	return &Server{
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	slog.Info("starting HTTP server", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(start).String(),
		)
	})
}

