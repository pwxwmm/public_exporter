// Author: mmwei3
// Email: mmwei3@iflytek.com
// Date: 2025-04-03
//
// Description:
// This package defines the ExporterService, which manages the initialization 
// and execution of the CollectorManager. It ensures background tasks are 
// properly started for Prometheus metric collection.

package service

import (
	"log"
	"public_exporter/config"
	"public_exporter/collector"
)

// ExporterService is the service layer that coordinates the CollectorManager.
type ExporterService struct {
	Config           *config.Config
	CollectorManager *collector.CollectorManager
}

// NewExporterService creates a new ExporterService.
func NewExporterService(cfg *config.Config, cm *collector.CollectorManager) *ExporterService {
	return &ExporterService{
		Config:           cfg,
		CollectorManager: cm,
	}
}

// Start starts the background tasks: collectors.
func (es *ExporterService) Start() error {
	log.Println("Starting exporter service...")
	
	// Start collector routines
	if err := es.CollectorManager.Start(); err != nil {
		return err
	}
	
	log.Println("Exporter service started successfully.")
	return nil
}

// Stop gracefully stops the exporter service.
func (es *ExporterService) Stop() {
	log.Println("Stopping exporter service...")
	es.CollectorManager.Stop()
	log.Println("Exporter service stopped.")
}
