// Package runner orchestrates application initialization and lifecycle.
package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"vista-app/internal/app/auth"
	"vista-app/internal/app/core"
	"vista-app/internal/app/ipam"
)

// Service orchestrates application initialization and execution.
type Service interface {
	Run(configPath string) error
}

// Container holds initialized application services.
type Container struct {
	Config  *core.Config
	Logger  core.LoggerService
	DB      core.DatabaseService
	JWTSvc  auth.JWTService
	IPAMSvc ipam.IPAMService
}

// serviceImpl is the concrete implementation of Service.
type serviceImpl struct {
	container *Container
}

// Run orchestrates the complete application lifecycle.
func (s *serviceImpl) Run(configPath string) error {
	// Step 1: Load configuration
	configLoader := core.NewConfigLoader()
	cfg, err := configLoader.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Step 2: Initialize logger
	logger := core.NewLoggerService()
	if err := logger.Init(cfg.Logging); err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	logger.Info("Application starting",
		zap.String("config_path", configPath))

	// Initialize container
	s.container = &Container{
		Config: cfg,
		Logger: logger,
	}

	// Step 3: Initialize database with retry logic
	logger.Info("Initializing database connection")
	db := core.NewDatabaseService()
	if err := db.Init(cfg.Database, logger); err != nil {
		logger.Error("Database initialization failed", zap.Error(err))
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer func() {
		logger.Info("Closing database connection")
		if err := db.CloseDB(); err != nil {
			logger.Error("Database close failed", zap.Error(err))
		}
	}()

	s.container.DB = db

	// Step 4: Run database migrations
	logger.Info("Running database migrations")
	if err := db.RunInitialMigration(
		&ipam.IPPool{},
		&ipam.IPLease{},
		&ipam.PortPool{},
		&ipam.PortLease{},
		&ipam.Audit{},
		&ipam.NodeCapability{},
	); err != nil {
		logger.Error("Database migration failed", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Step 5: Initialize JWT service
	logger.Info("Initializing JWT service")
	jwtSvc := auth.NewJWTService()
	jwtSvc.Init(cfg.Auth)
	s.container.JWTSvc = jwtSvc

	// Step 6: Initialize IPAM service
	logger.Info("Initializing IPAM service")
	ipamSvc := ipam.NewIPAMService()
	ipamSvc.Init(db.GetDB(), cfg.IPAM)
	s.container.IPAMSvc = ipamSvc

	logger.Info("All core services initialized successfully")

	// Step 7: Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle shutdown signals
	go func() {
		sig := <-sigChan
		logger.Info("Shutdown signal received",
			zap.String("signal", sig.String()))
		logger.Info("Application shutdown initiated")
		cancel()
	}()

	// Step 8: HTTP server placeholder
	logger.Info("HTTP server initialization placeholder")
	logger.Info("API handlers not yet implemented")

	// TODO: When API handlers ready:
	// router := api.NewRouter()
	// handler := router.Setup(logger, jwtSvc, ipamSvc)
	// server := &http.Server{Addr: ":8080", Handler: handler}
	// go server.ListenAndServe()
	// defer server.Shutdown(context.Background())

	// Wait for shutdown signal
	logger.Info("Waiting for shutdown signal (Ctrl+C)")
	<-ctx.Done()

	logger.Info("Application shutdown complete")
	return nil
}

// NewService creates a new Service instance.
func NewService() Service {
	return &serviceImpl{}
}
