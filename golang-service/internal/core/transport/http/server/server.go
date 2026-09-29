package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	core_logger "github.com/m0dris/hackathon/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux    *http.ServeMux
	config Config
	log    *core_logger.Logger
}

func NewHTTPServer(config Config, log *core_logger.Logger) *HTTPServer {
	return &HTTPServer{
		mux:    http.NewServeMux(),
		config: config,
		log:    log,
	}
}

func (s *HTTPServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              s.config.Addr,
		Handler:           s.logRequests(s.mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		s.log.Warn("start HTTP server", zap.String("addr", s.config.Addr))

		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
	case <-ctx.Done():
		s.log.Warn("shutdown HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		s.log.Warn("HTTP server stopped")
	}

	return nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (s *HTTPServer) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		s.log.Debug(
			"HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rec.status),
			zap.Duration("latency", time.Since(start)),
		)
	})
}
