// Package api provides the HTTP API implementation for the VPN Automation System.
package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"D:/vista-app/internal/app/auth"
	"D:/vista-app/internal/app/core"
	"D:/vista-app/internal/app/ipam"
)

// RouterService defines methods for API routing and serving.
type RouterService interface {
	// Init initializes the router with middleware and handlers.
	Init() error
	// Serve starts the HTTP server.
	Serve() error
}

// Router implements RouterService for the application.
type Router struct {
	cfg        core.Config
	logger     core.LoggerService
	jwtSvc     auth.JWTService
	ipamSvc    ipam.Service
	router     *chi.Mux
	handlers   *APICallbacks
	serverPort string
}

// NewRouter creates a new API router with dependencies.
func NewRouter(
	cfg core.Config,
	logger core.LoggerService,
	jwtSvc auth.JWTService,
	ipamSvc ipam.Service,
) RouterService {
	return &Router{
		cfg:     cfg,
		logger:  logger,
		jwtSvc:  jwtSvc,
		ipamSvc: ipamSvc,
		router:  chi.NewRouter(),
	}
}

// Init initializes the router with middleware and handlers.
func (r *Router) Init() error {
	r.serverPort = r.cfg.GetString("api.port")
	if r.serverPort == "" {
		r.serverPort = "8080" // Default port from config fallback
	}

	r.handlers = &APICallbacks{
		cfg:     r.cfg,
		logger:  r.logger,
		jwtSvc:  r.jwtSvc,
		ipamSvc: r.ipamSvc,
	}

	r.setupMiddleware()
	r.mountHandlers()

	r.logger.Info("API router initialized", "port", r.serverPort)
	return nil
}

// Serve starts the HTTP server.
func (r *Router) Serve() error {
	server := &http.Server{
		Addr:         ":" + r.serverPort,
		Handler:      r.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	r.logger.Info("API server starting", "port", r.serverPort)
	return server.ListenAndServe()
}

// setupMiddleware configures all middleware for the router.
func (r *Router) setupMiddleware() {
	r.router.Use(middleware.RequestID)
	r.router.Use(r.loggerMiddleware)
	r.router.Use(middleware.Recoverer)
	r.router.Use(middleware.RealIP)
	r.router.Use(r.corsMiddleware)
	r.router.Use(middleware.Timeout(30 * time.Second))
}

// mountHandlers maps routes to their respective handlers.
func (r *Router) mountHandlers() {
	// Public routes
	r.router.Group(func(public chi.Router) {
		public.Get("/api/v1/health", r.handlers.HealthCheck)
		public.Post("/api/v1/auth/login", r.handlers.Login)
	})

	// Protected routes
	r.router.Group(func(protected chi.Router) {
		protected.Use(r.authMiddleware)

		// IPAM endpoints
		protected.Post("/api/v1/ipam/reserve-ip", r.handlers.ReserveIP)
		protected.Post("/api/v1/ipam/commit-ip", r.handlers.CommitIP)
		protected.Post("/api/v1/ipam/recycle-node/{nodeID}", r.handlers.RecycleNode)
	})
}

// loggerMiddleware logs HTTP request details.
func (r *Router) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, req.ProtoMajor)

		defer func() {
			r.logger.Info("API request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start).String(),
				"size", ww.BytesWritten(),
			)
		}()

		next.ServeHTTP(ww, req)
	})
}

// corsMiddleware handles Cross-Origin Resource Sharing.
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := r.cfg.GetString("api.cors.origin")
		if origin == "" {
			origin = "*" // Default from config fallback
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", 
			"GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", 
			"Accept, Content-Type, Content-Length, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}