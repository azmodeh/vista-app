// Package core provides core application services.
package core

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigLoaderService defines the interface for loading configuration.
type ConfigLoaderService interface {
	Load(path string) (*Config, error)
}

// Config is the root configuration structure for the entire app.
type Config struct {
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Auth     AuthConfig     `yaml:"auth" json:"auth"`
	Network  NetworkConfig  `yaml:"network" json:"network"`
	IPAM     IPAMConfig     `yaml:"ipam" json:"ipam"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	DSN         string `yaml:"dsn" json:"dsn"`
	MaxOpen     int    `yaml:"max_open" json:"max_open"`
	MaxIdle     int    `yaml:"max_idle" json:"max_idle"`
	MaxLifetime string `yaml:"max_lifetime" json:"max_lifetime"`
}

// AuthConfig holds authentication and JWT settings.
type AuthConfig struct {
	JWTSigningKey string `yaml:"jwt_signing_key" json:"-"`
	TokenExpiry   string `yaml:"token_expiry" json:"token_expiry"`
	Issuer        string `yaml:"issuer" json:"issuer"`
}

// NetworkConfig contains infrastructure endpoints and policies.
type NetworkConfig struct {
	GeoIPPrimaryURL   string `yaml:"geoip_primary" json:"geoip_primary"`
	GeoIPFallbackURL  string `yaml:"geoip_fallback" json:"geoip_fallback"`
	PingPrimaryNode   string `yaml:"ping_primary" json:"ping_primary"`
	PingFallbackNode  string `yaml:"ping_fallback" json:"ping_fallback"`
}

// IPAMConfig defines IP and Port allocation pools and policies.
type IPAMConfig struct {
	IPPools       []IPPool    `yaml:"ip_pools" json:"ip_pools"`
	PortRanges    []PortRange `yaml:"port_ranges" json:"port_ranges"`
	LeaseDuration string      `yaml:"lease_duration" json:"lease_duration"`
}

// IPPool defines a CIDR block allocation pool.
type IPPool struct {
	CIDR     string `yaml:"cidr" json:"cidr"`
	NodeID   string `yaml:"node_id" json:"node_id"`
	IsGlobal bool   `yaml:"is_global" json:"is_global"`
}

// PortRange defines a port allocation range.
type PortRange struct {
	Start     int    `yaml:"start" json:"start"`
	End       int    `yaml:"end" json:"end"`
	Protocol  string `yaml:"protocol" json:"protocol"`
	Blacklist []int  `yaml:"blacklist" json:"blacklist"`
}

// configLoaderImpl is the concrete implementation.
type configLoaderImpl struct{}

// Global singleton configuration instance.
var (
	configuration *Config
	loadedOnce    bool
)

// Load reads the YAML configuration file and returns Config.
// Secrets (JWT key) are overridden by environment variables.
func (c *configLoaderImpl) Load(path string) (*Config, error) {
	// Read YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal YAML into Config
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override secrets from environment variables
	if err := c.loadSecretsFromEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	// Validate critical fields
	if err := c.validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Store in singleton
	configuration = &cfg
	loadedOnce = true

	return &cfg, nil
}

// loadSecretsFromEnv overrides config with environment variables.
func (c *configLoaderImpl) loadSecretsFromEnv(cfg *Config) error {
	// JWT Signing Key from environment (required for production)
	if jwtKey := os.Getenv("JWT_SIGNING_KEY"); jwtKey != "" {
		cfg.Auth.JWTSigningKey = jwtKey
	}

	// Database DSN can also be overridden
	if dbDSN := os.Getenv("DATABASE_DSN"); dbDSN != "" {
		cfg.Database.DSN = dbDSN
	}

	return nil
}

// validate checks that critical configuration fields are set.
func (c *configLoaderImpl) validate(cfg *Config) error {
	// Database DSN must be present
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database DSN cannot be empty")
	}

	// JWT signing key must be present
	if cfg.Auth.JWTSigningKey == "" {
		return fmt.Errorf("JWT signing key cannot be empty")
	}

	// Logging level must be valid
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true,
		"error": true, "fatal": true,
	}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", cfg.Logging.Level)
	}

	// Network endpoints validation
	if cfg.Network.GeoIPPrimaryURL == "" {
		return fmt.Errorf("GeoIP primary URL cannot be empty")
	}
	if cfg.Network.PingPrimaryNode == "" {
		return fmt.Errorf("Ping primary node cannot be empty")
	}

	// IPAM validation
	if len(cfg.IPAM.IPPools) == 0 {
		return fmt.Errorf("at least one IP pool must be defined")
	}
	if len(cfg.IPAM.PortRanges) == 0 {
		return fmt.Errorf("at least one port range must be defined")
	}

	return nil
}

// NewConfigLoader creates a new ConfigLoaderService instance.
func NewConfigLoader() ConfigLoaderService {
	return &configLoaderImpl{}
}

// LoadConfig is a convenience function to load configuration.
func LoadConfig(path string) (*Config, error) {
	loader := NewConfigLoader()
	return loader.Load(path)
}

// MustLoadConfig loads config and panics on error.
func MustLoadConfig(path string) *Config {
	cfg, err := LoadConfig(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// GetConfig returns the singleton configuration instance.
// Returns nil if configuration has not been loaded yet.
func GetConfig() *Config {
	return configuration
}

// GetTokenExpiryDuration parses the token expiry string.
func (c *Config) GetTokenExpiryDuration() (time.Duration, error) {
	return time.ParseDuration(c.Auth.TokenExpiry)
}

// GetLeaseDuration parses the IPAM lease duration string.
func (c *Config) GetLeaseDuration() (time.Duration, error) {
	return time.ParseDuration(c.IPAM.LeaseDuration)
}

// GetMaxLifetimeDuration parses the DB max lifetime string.
func (c *Config) GetMaxLifetimeDuration() (time.Duration, error) {
	if c.Database.MaxLifetime == "" {
		return 0, nil // No limit
	}
	return time.ParseDuration(c.Database.MaxLifetime)
}

// IsLoaded returns whether configuration has been loaded.
func IsLoaded() bool {
	return loadedOnce
}
