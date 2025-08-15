// Author: mmwei3
// Email: mmwei3@iflytek.com
// Date: 2025-04-03
//
// Description:
// This package provides configuration management for a monitoring system.
// It loads YAML configuration, sets default values, and configures logging with rotation support.
// The configuration supports multiple clusters and collectors with customizable intervals and timeouts.

package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/lestrrat-go/file-rotatelogs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Config holds the global configuration.
type Config struct {
	Global  GlobalConfig              `yaml:"global"`
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

// GlobalConfig holds global configuration settings.
type GlobalConfig struct {
	LogFile             string `yaml:"log_file"`
	LogLevel            string `yaml:"log_level"`
	LogMaxAge           int    `yaml:"log_max_age"`
	LogRotationTime     int    `yaml:"log_rotation_time"`
	DefaultScrapeInterval int  `yaml:"default_scrape_interval"`
	HTTPPort            int    `yaml:"http_port"`
	HTTPTimeout         int    `yaml:"http_timeout"`
}

// ClusterConfig represents the configuration for a cluster.
type ClusterConfig struct {
	Enabled    bool                       `yaml:"enabled"`
	Collectors map[string]CollectorConfig `yaml:"collectors"`
}

// CollectorConfig holds the configuration for a collector.
type CollectorConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Interval   int    `yaml:"interval"`
	Timeout    int    `yaml:"timeout"`
	ScriptPath string `yaml:"script_path"`
	ScriptType string `yaml:"script_type"`
}

// LoadConfig loads the YAML configuration from the specified path.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default values and validate
	if err := cfg.setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}
	
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration fields.
func (c *Config) setDefaults() error {
	// Global defaults
	if c.Global.LogLevel == "" {
		c.Global.LogLevel = "info"
	}
	if c.Global.LogMaxAge == 0 {
		c.Global.LogMaxAge = 7 // Default: 7 days
	}
	if c.Global.LogRotationTime == 0 {
		c.Global.LogRotationTime = 24 // Default: 24 hours
	}
	if c.Global.DefaultScrapeInterval == 0 {
		c.Global.DefaultScrapeInterval = 60 // Default: 60 seconds
	}
	if c.Global.HTTPPort == 0 {
		c.Global.HTTPPort = 5535 // Default: 5535
	}
	if c.Global.HTTPTimeout == 0 {
		c.Global.HTTPTimeout = 30 // Default: 30 seconds
	}
	
	// Collector defaults
	for clusterName, clusterCfg := range c.Clusters {
		for collectorName, collectorCfg := range clusterCfg.Collectors {
			if collectorCfg.Interval == 0 {
				collectorCfg.Interval = c.Global.DefaultScrapeInterval
			}
			if collectorCfg.Timeout == 0 {
				collectorCfg.Timeout = 30 // Default: 30 seconds
			}
			// Update the collector config in the map
			clusterCfg.Collectors[collectorName] = collectorCfg
		}
		c.Clusters[clusterName] = clusterCfg
	}
	
	return nil
}

// validate validates the configuration.
func (c *Config) validate() error {
	// Validate global settings
	if c.Global.LogFile == "" {
		return fmt.Errorf("global.log_file is required")
	}
	
	if c.Global.LogMaxAge <= 0 {
		return fmt.Errorf("global.log_max_age must be positive, got %d", c.Global.LogMaxAge)
	}
	
	if c.Global.LogRotationTime <= 0 {
		return fmt.Errorf("global.log_rotation_time must be positive, got %d", c.Global.LogRotationTime)
	}
	
	if c.Global.DefaultScrapeInterval <= 0 {
		return fmt.Errorf("global.default_scrape_interval must be positive, got %d", c.Global.DefaultScrapeInterval)
	}
	
	if c.Global.HTTPPort <= 0 || c.Global.HTTPPort > 65535 {
		return fmt.Errorf("global.http_port must be between 1 and 65535, got %d", c.Global.HTTPPort)
	}
	
	if c.Global.HTTPTimeout <= 0 {
		return fmt.Errorf("global.http_timeout must be positive, got %d", c.Global.HTTPTimeout)
	}
	
	// Validate clusters and collectors
	if len(c.Clusters) == 0 {
		return fmt.Errorf("at least one cluster must be configured")
	}
	
	for clusterName, clusterCfg := range c.Clusters {
		if clusterCfg.Enabled {
			if len(clusterCfg.Collectors) == 0 {
				return fmt.Errorf("cluster %s has no collectors configured", clusterName)
			}
			
			for collectorName, collectorCfg := range clusterCfg.Collectors {
				if collectorCfg.Enabled {
					if err := validateCollectorConfig(collectorName, collectorCfg); err != nil {
						return fmt.Errorf("cluster %s, collector %s: %w", clusterName, collectorName, err)
					}
				}
			}
		}
	}
	
	return nil
}

// validateCollectorConfig validates individual collector configuration.
func validateCollectorConfig(name string, cfg CollectorConfig) error {
	if cfg.Interval <= 0 {
		return fmt.Errorf("interval must be positive, got %d", cfg.Interval)
	}
	
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %d", cfg.Timeout)
	}
	
	if cfg.ScriptPath == "" {
		return fmt.Errorf("script_path cannot be empty")
	}
	
	if cfg.ScriptType == "" {
		return fmt.Errorf("script_type cannot be empty")
	}
	
	// Validate script type
	validTypes := map[string]bool{
		"python":  true,
		"python2": true,
		"python3": true,
		"shell":   true,
	}
	
	if !validTypes[cfg.ScriptType] {
		return fmt.Errorf("unsupported script_type: %s, supported types: python, python2, python3, shell", cfg.ScriptType)
	}
	
	return nil
}

// SetupLogging configures the log output with rotation.
func SetupLogging(logFile string, logLevel string, logMaxAge int, logRotationTime int) {
	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel // Default to info if parsing fails
		logrus.Warnf("Invalid log level %s, defaulting to info", logLevel)
	}
	logrus.SetLevel(level)

	// Create log directory if not exists
	dir := filepath.Dir(logFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logrus.Fatalf("Failed to create log directory %s: %v", dir, err)
		}
	}

	// Configure log rotation
	logWriter, err := rotatelogs.New(
		logFile+".%Y%m%d%H",
		rotatelogs.WithMaxAge(time.Duration(logMaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(logRotationTime)*time.Hour),
	)
	if err != nil {
		logrus.Fatalf("Failed to set up log rotation: %v", err)
	}

	logrus.SetOutput(logWriter)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	logrus.Infof("Logging configured: level=%s, file=%s, max_age=%dd, rotation_time=%dh", 
		level, logFile, logMaxAge, logRotationTime)
}
