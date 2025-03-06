package main

import (
	"fmt"
	"net/http"
	"os"

	"publice_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func main() {
	deviceID := 0
	if val, ok := os.LookupEnv("DEVICE_ID"); ok {
		fmt.Sscanf(val, "%d", &deviceID)
	}

	// 注册采集器
	hccnCollector := collector.NewHCCNCollector(deviceID)
	dockerCollector := collector.NewDockerRuntimeCollector()

	prometheus.MustRegister(hccnCollector)
	prometheus.MustRegister(dockerCollector)

	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting publice_exporter on :5535")
	log.Fatal(http.ListenAndServe(":5535", nil))
}