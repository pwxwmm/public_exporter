package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"public_exporter/config"
	"public_exporter/collector"
	"public_exporter/service"
	"strings"
	"syscall"
	"time"
)

const (
	Author = "mmwei3"
	Email  = "mmwei3@iflytek.com, 1300042631@qq.com"
	Date   = "2025-03-28"
	Version = "1.0.0"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config.file", "/app/config/config.yaml", "Path to configuration file")
	flag.Parse()
}

func main() {
	fmt.Println("=====================================")
	fmt.Println("         Public Exporter             ")
	fmt.Printf("         Version: %s                \n", Version)
	fmt.Println("=====================================")
	fmt.Printf("Author: %s\nEmail: %s\nDate: %s\n", Author, Email, Date)

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	if err := setupLogging(cfg); err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		os.Exit(1)
	}

	log.Println("Starting public_exporter...")

	// Create and start services
	collectorManager := collector.NewCollectorManager(cfg)
	exporterService := service.NewExporterService(cfg, collectorManager)
	
	if err := exporterService.Start(); err != nil {
		log.Fatalf("Failed to start exporter service: %v", err)
	}

	// Setup HTTP server
	server := setupHTTPServer(cfg, collectorManager)
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		
		log.Println("Received shutdown signal, starting graceful shutdown...")
		cancel()
		
		// Give some time for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
		
		exporterService.Stop()
		log.Println("Graceful shutdown completed")
		os.Exit(0)
	}()

	port := cfg.Global.HTTPPort
	if port == 0 {
		port = 5535
	}
	log.Printf("Exporter is running on http://0.0.0.0:%d/metrics", port)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupLogging(cfg *config.Config) error {
	// Setup logrus logging with rotation
	config.SetupLogging(
		cfg.Global.LogFile,
		cfg.Global.LogLevel,
		cfg.Global.LogMaxAge,
		cfg.Global.LogRotationTime,
	)
	return nil
}

func setupHTTPServer(cfg *config.Config, collectorManager *collector.CollectorManager) *http.Server {
	mux := http.NewServeMux()
	
	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var outputs []string
		globalHealthy := 1

		// Get collector outputs
		collectorOutputs := collectorManager.GetOutputs()
		outputs = append(outputs, collectorOutputs...)

		// Get health status
		healthStatus := collectorManager.GetHealthStatus()
		for key, health := range healthStatus {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				cluster, name := parts[0], parts[1]
				if health == 0 {
					globalHealthy = 0
				}
				outputs = append(outputs, fmt.Sprintf(`collector_health_status{cluster="%s", collector="%s"} %d`, cluster, name, health))
			}
		}

		// Add exporter health status
		outputs = append(outputs, `# HELP exporter_health_status Global health status of the exporter`)
		outputs = append(outputs, `# TYPE exporter_health_status gauge`)
		outputs = append(outputs, fmt.Sprintf("exporter_health_status %d", globalHealthy))
		
		// Add collector count metric
		outputs = append(outputs, `# HELP collector_count Total number of active collectors`)
		outputs = append(outputs, `# TYPE collector_count gauge`)
		outputs = append(outputs, fmt.Sprintf("collector_count %d", collectorManager.GetCollectorCount()))

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Join(outputs, "\n")))
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		globalHealthy := "ok"
		collectorStatuses := make(map[string]string)

		healthStatus := collectorManager.GetHealthStatus()
		for key, health := range healthStatus {
			if health == 0 {
				collectorStatuses[key] = "failed"
				globalHealthy = "failed"
			} else {
				collectorStatuses[key] = "ok"
			}
		}

		output := fmt.Sprintf(`{"status":"%s", "collectors":{`, globalHealthy)
		var items []string
		for k, v := range collectorStatuses {
			items = append(items, fmt.Sprintf(`"%s":"%s"`, k, v))
		}
		output += strings.Join(items, ",") + `}}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(output))
	})

	// Root endpoint with basic info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Public Exporter</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>Public Exporter</h1>
    <p>Version: %s</p>
    <p>Author: %s</p>
    <p>Email: %s</p>
    <ul>
        <li><a href="/metrics">Metrics</a> - Prometheus metrics endpoint</li>
        <li><a href="/health">Health</a> - Health check endpoint</li>
    </ul>
</body>
</html>`, Version, Author, Email)
		w.Write([]byte(html))
	})

	// Get port from config or use default
	port := cfg.Global.HTTPPort
	if port == 0 {
		port = 5535
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Global.HTTPTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Global.HTTPTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Global.HTTPTimeout*2) * time.Second,
	}
}
