package config

import (
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
	Global struct {
		LogFile          string `yaml:"log_file"`
		LogLevel         string `yaml:"log_level"`
		LogMaxAge        int    `yaml:"log_max_age"`
		LogRotationTime  int    `yaml:"log_rotation_time"`
		DefaultScrapeInterval int `yaml:"default_scrape_interval"`
	} `yaml:"global"`
	Clusters map[string]ClusterConfig `yaml:"clusters"`
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
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set default values if not provided
	if cfg.Global.LogLevel == "" {
		cfg.Global.LogLevel = "info"
	}
	if cfg.Global.LogMaxAge == 0 {
		cfg.Global.LogMaxAge = 7 // Default: 7 days
	}
	if cfg.Global.LogRotationTime == 0 {
		cfg.Global.LogRotationTime = 24 // Default: 24 hours
	}

	return &cfg, nil
}

// SetupLogging configures the log output with rotation.
func SetupLogging(logFile string, logLevel string, logMaxAge int, logRotationTime int) {
	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel // Default to info if parsing fails
	}
	logrus.SetLevel(level)

	// Create log directory if not exists
	dir := filepath.Dir(logFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
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
}
