// Package core provides core application services including logging.
package core

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerService defines the methods required for structured logging.
// It acts as a wrapper around a Zap logger instance.
type LoggerService interface {
	Init(cfg LoggingConfig) error
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) *zap.SugaredLogger
}

// LoggingConfig holds configuration settings for the logger.
// It is loaded by config_loader.go from YAML.
type LoggingConfig struct {
	Level      string // e.g., "debug", "info", "warn"
	Encoding   string // e.g., "json", "console"
	OutputPath string // e.g., "stdout", "stderr", or file path
}

// loggerImpl is the concrete implementation of LoggerService.
type loggerImpl struct {
	logger *zap.Logger
}

// Global singleton logger instance
var (
	instance LoggerService
	once     sync.Once
)

// Init initializes the logger with the provided configuration.
// This method should be called once at application startup.
// It uses sync.Once to ensure singleton initialization.
func (l *loggerImpl) Init(cfg LoggingConfig) error {
	var initErr error
	once.Do(func() {
		level := parseLogLevel(cfg.Level)
		zapCfg := buildZapConfig(level, cfg.Encoding, cfg.OutputPath)

		logger, err := zapCfg.Build()
		if err != nil {
			initErr = err
			return
		}

		l.logger = logger
		instance = l
	})
	return initErr
}

// Debug logs a debug-level message with optional fields.
func (l *loggerImpl) Debug(msg string, fields ...zap.Field) {
	if l.logger != nil {
		l.logger.Debug(msg, fields...)
	}
}

// Info logs an info-level message with optional fields.
func (l *loggerImpl) Info(msg string, fields ...zap.Field) {
	if l.logger != nil {
		l.logger.Info(msg, fields...)
	}
}

// Warn logs a warning-level message with optional fields.
func (l *loggerImpl) Warn(msg string, fields ...zap.Field) {
	if l.logger != nil {
		l.logger.Warn(msg, fields...)
	}
}

// Error logs an error-level message with optional fields.
func (l *loggerImpl) Error(msg string, fields ...zap.Field) {
	if l.logger != nil {
		l.logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal-level message and terminates the application.
func (l *loggerImpl) Fatal(msg string, fields ...zap.Field) {
	if l.logger != nil {
		l.logger.Fatal(msg, fields...)
	}
}

// With creates a child logger with additional context fields.
func (l *loggerImpl) With(fields ...zap.Field) *zap.SugaredLogger {
	if l.logger != nil {
		return l.logger.With(fields...).Sugar()
	}
	return nil
}

// parseLogLevel converts a string log level to zapcore.Level.
// Defaults to InfoLevel if the string is invalid.
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// buildZapConfig constructs a zap.Config based on provided settings.
func buildZapConfig(
	level zapcore.Level,
	encoding string,
	outputPath string,
) zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	// Set encoding: default to "console" if invalid
	if encoding == "json" || encoding == "console" {
		cfg.Encoding = encoding
	} else {
		cfg.Encoding = "console"
	}

	// Set output path: default to "stdout" if empty
	if outputPath == "" {
		outputPath = "stdout"
	}
	cfg.OutputPaths = []string{outputPath}
	cfg.ErrorOutputPaths = []string{"stderr"}

	return cfg
}

// NewLoggerService creates a new LoggerService instance.
func NewLoggerService() LoggerService {
	return &loggerImpl{}
}

// Log returns the global singleton logger instance.
// Must call Init before using this accessor.
func Log() LoggerService {
	if instance == nil {
		// Initialize with default config if not already done
		defaultLogger := &loggerImpl{}
		_ = defaultLogger.Init(LoggingConfig{
			Level:      "info",
			Encoding:   "console",
			OutputPath: "stdout",
		})
	}
	return instance
}
