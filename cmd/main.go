package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"io"
	"net/http" 

	"public_exporter/config"
	"public_exporter/collector"
	"public_exporter/scheduler"
	"public_exporter/service"
)

// Author information constants.
const (
	Author = "mmwei3"
	Email  = "mmwei3@iflytek.com, 1300042631@qq.com"
	Date   = "2025-03-28"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config.file", "/app/config/config.yaml", "Path to configuration file")
	flag.Parse()
}

func main() {
	// Print author information.
	fmt.Println("=====================================")
	fmt.Println("         Public Exporter             ")
	fmt.Println("=====================================")
	fmt.Printf("Author: %s\nEmail: %s\nDate: %s\n", Author, Email, Date)

	// Load configuration.
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup logging - log to both stdout and the log file.
	logFile, err := os.OpenFile(cfg.Global.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Create a multi-writer to log to both stdout and the log file.
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.Println("Starting public_exporter...")

	// Initialize CollectorManager, Scheduler, and ExporterService.
	collectorManager := collector.NewCollectorManager(cfg)
	sched := scheduler.NewScheduler()
	exporterService := service.NewExporterService(cfg, collectorManager, sched)
	exporterService.Start()

	// Create a custom HTTP handler for the /metrics endpoint.
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var outputs []string
		collector.CollectorOutputs.Range(func(key, value interface{}) bool {
			outputs = append(outputs, fmt.Sprintf("%s", value))
			return true
		})
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Join(outputs, "\n")))
	})

	// Add a simple health check endpoint.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Start the HTTP server.
	port := 5535
	log.Printf("Exporter is running on http://0.0.0.0:%d/metrics", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
