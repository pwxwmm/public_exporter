package service

import (
	"log"
	"public_exporter/config"
	"public_exporter/collector"
	"public_exporter/scheduler"
)

// ExporterService is the service layer that coordinates the CollectorManager and Scheduler.
type ExporterService struct {
	Config           *config.Config
	CollectorManager *collector.CollectorManager
	Scheduler        *scheduler.Scheduler
}

// NewExporterService creates a new ExporterService.
func NewExporterService(cfg *config.Config, cm *collector.CollectorManager, s *scheduler.Scheduler) *ExporterService {
	return &ExporterService{
		Config:           cfg,
		CollectorManager: cm,
		Scheduler:        s,
	}
}

// Start starts the background tasks: collectors and scheduler.
func (es *ExporterService) Start() {
	// Start collector routines.
	go es.CollectorManager.RegisterAll()
	// Start scheduler routines.
	go es.Scheduler.StartScheduler(es.Config)
	log.Println("Exporter service started.")
}
