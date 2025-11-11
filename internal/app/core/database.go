// Package core provides foundational services for the application.
package core

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseService defines methods for database connectivity and migrations.
type DatabaseService interface {
	Init(cfg DatabaseConfig, log LoggerService) error
	GetDB() *gorm.DB
	RunInitialMigration(models ...interface{}) error
	CloseDB() error
}

// databaseImpl is the concrete implementation of DatabaseService.
type databaseImpl struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	logger LoggerService
	config DatabaseConfig
}

var dbService DatabaseService

// Init initializes the database connection with retry logic.
func (d *databaseImpl) Init(
	cfg DatabaseConfig,
	log LoggerService,
) error {
	d.config = cfg
	d.logger = log

	var err error
	maxRetries := 3
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		d.logger.Info("Attempting database connection",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries))

		d.db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
			Logger: nil, // Disable GORM's default logger
		})

		if err != nil {
			d.logger.Warn("Database connection failed",
				zap.Int("attempt", attempt),
				zap.Error(err))

			if attempt < maxRetries {
				d.logger.Info("Retrying connection",
					zap.Duration("delay", retryDelay))
				time.Sleep(retryDelay)
				continue
			}

			d.logger.Error("All database connection attempts failed",
				zap.Error(err))
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		// Get underlying SQL DB for connection pool settings
		d.sqlDB, err = d.db.DB()
		if err != nil {
			d.logger.Error("Failed to get SQL DB instance",
				zap.Error(err))
			return fmt.Errorf("failed to get SQL DB: %w", err)
		}

		// Apply connection pool settings
		d.sqlDB.SetMaxOpenConns(cfg.MaxOpen)
		d.sqlDB.SetMaxIdleConns(cfg.MaxIdle)

		d.logger.Info("Database connected successfully",
			zap.Int("max_open_conns", cfg.MaxOpen),
			zap.Int("max_idle_conns", cfg.MaxIdle))

		dbService = d
		return nil
	}

	return fmt.Errorf("unexpected database connection failure")
}

// GetDB returns the GORM database instance.
func (d *databaseImpl) GetDB() *gorm.DB {
	return d.db
}

// RunInitialMigration runs GORM AutoMigrate for provided models.
func (d *databaseImpl) RunInitialMigration(
	models ...interface{},
) error {
	if len(models) == 0 {
		d.logger.Warn("No models provided for migration")
		return nil
	}

	d.logger.Info("Starting database migration",
		zap.Int("model_count", len(models)))

	err := d.db.AutoMigrate(models...)
	if err != nil {
		d.logger.Error("Database migration failed", zap.Error(err))
		return fmt.Errorf("migration failed: %w", err)
	}

	d.logger.Info("Database migration completed successfully")
	return nil
}

// CloseDB closes the database connection gracefully.
func (d *databaseImpl) CloseDB() error {
	if d.sqlDB == nil {
		d.logger.Warn("No database connection to close")
		return nil
	}

	d.logger.Info("Closing database connection")

	if err := d.sqlDB.Close(); err != nil {
		d.logger.Error("Failed to close database", zap.Error(err))
		return fmt.Errorf("failed to close database: %w", err)
	}

	d.logger.Info("Database connection closed successfully")
	return nil
}

// NewDatabaseService creates a new DatabaseService instance.
func NewDatabaseService() DatabaseService {
	return &databaseImpl{}
}

// GetDatabaseService returns the global singleton database service.
func GetDatabaseService() DatabaseService {
	return dbService
}
