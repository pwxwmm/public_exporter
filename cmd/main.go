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
	fmt.Println("=====================================")
	fmt.Println("         Public Exporter             ")
	fmt.Println("=====================================")
	fmt.Printf("Author: %s\nEmail: %s\nDate: %s\n", Author, Email, Date)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(cfg.Global.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	log.Println("Starting public_exporter...")

	collectorManager := collector.NewCollectorManager(cfg)
	sched := scheduler.NewScheduler()
	exporterService := service.NewExporterService(cfg, collectorManager, sched)
	exporterService.Start()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var outputs []string
		globalHealthy := 1

		collector.CollectorOutputs.Range(func(_, value interface{}) bool {
			outputs = append(outputs, fmt.Sprintf("%s", value))
			return true
		})

		collector.CollectorHealth.Range(func(key, value interface{}) bool {
			parts := strings.Split(key.(string), ":")
			cluster, name := parts[0], parts[1]
			health := value.(int)
			if health == 0 {
				globalHealthy = 0
			}
			outputs = append(outputs, fmt.Sprintf(`collector_health_status{cluster="%s", collector="%s"} %d`, cluster, name, health))
			return true
		})

		outputs = append(outputs, `# HELP exporter_health_status Global health status of the exporter`)
		outputs = append(outputs, `# TYPE exporter_health_status gauge`)
		outputs = append(outputs, fmt.Sprintf("exporter_health_status %d", globalHealthy))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Join(outputs, "\n")))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		globalHealthy := "ok"
		collectorStatuses := make(map[string]string)

		collector.CollectorHealth.Range(func(key, value interface{}) bool {
			name := key.(string)
			health := value.(int)
			if health == 0 {
				collectorStatuses[name] = "failed"
				globalHealthy = "failed"
			} else {
				collectorStatuses[name] = "ok"
			}
			return true
		})

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

	port := 5535
	log.Printf("Exporter is running on http://0.0.0.0:%d/metrics", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
