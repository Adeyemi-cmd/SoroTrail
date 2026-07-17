// Package api serves stored events over HTTP. Endpoints are documented in
// the README's API reference.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/khaylebfortune/sorotrail/internal/rpc"
	"github.com/khaylebfortune/sorotrail/internal/store"
)

// Server holds the API's dependencies.
type Server struct {
	store store.Store
	rpc   rpc.Client
	log   *slog.Logger
}

// New builds the API server. rpcClient is only used by /health.
func New(st store.Store, rpcClient rpc.Client, log *slog.Logger) *Server {
	return &Server{store: st, rpc: rpcClient, log: log}
}

// Router returns the HTTP handler with all routes mounted.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(s.requestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", s.handleHealth)
	r.Get("/events", s.handleListEvents)
	r.Get("/events/{id}", s.handleGetEvent)
	r.Get("/contracts/{id}/events", s.handleContractEvents)
	r.Get("/stats", s.handleStats)

	// contributors: new read endpoints go here. Anything that writes (e.g.
	// managing watched contracts at runtime) should come with auth first.
	return r
}

func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		s.log.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"remote", r.RemoteAddr,
		)
	})
}
