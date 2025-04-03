// Author: mmwei3
// Email: mmwei3@iflytek.com
// Date: 2025-04-03
//
// Description:
// This package manages data collectors that execute external scripts periodically. 
// It supports Python and Shell scripts, ensuring their outputs are stored and exposed as metrics.
// The collected data is formatted for Prometheus compatibility and stored in a sync map.

package collector

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"public_exporter/config"
	"sync"
	"time"
)

// CollectorOutputs stores the latest output for each collector.
// The key format is "clusterName:collectorName".
var CollectorOutputs sync.Map

// ScriptExecutor is responsible for executing external scripts.
type ScriptExecutor struct {
	mu sync.Mutex
}

// ExecuteScript executes the specified script (Python or Shell) with a timeout.
// It returns the output as a string, the execution timestamp, or an error.
func (se *ScriptExecutor) ExecuteScript(scriptPath, scriptType string, timeout int) (string, string, error) {
	se.mu.Lock()
	defer se.mu.Unlock()

	execTime := time.Now().Format("2006-01-02 15:04:05.000")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if scriptType == "python" {
		cmd = exec.CommandContext(ctx, "python3", scriptPath)
	} else if scriptType == "shell" {
		cmd = exec.CommandContext(ctx, "bash", scriptPath)
	} else {
		return "", "", fmt.Errorf("unsupported script type: %s", scriptType)
	}

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", "", fmt.Errorf("script timed out")
	}
	if err != nil {
		return "", "", fmt.Errorf("error executing script: %v", err)
	}

	// In this project, we assume the output is already in the desired text format.
	// (No dynamic registration into Prometheus is needed.)
	return string(output), execTime, nil
}

// CollectorManager manages collectors by periodically executing their scripts.
type CollectorManager struct {
	Config         *config.Config
	ScriptExecutor *ScriptExecutor
}

// NewCollectorManager initializes a new CollectorManager.
func NewCollectorManager(cfg *config.Config) *CollectorManager {
	return &CollectorManager{
		Config:         cfg,
		ScriptExecutor: &ScriptExecutor{},
	}
}

// RegisterAll iterates over all enabled clusters and collectors,
// starting a goroutine for each collector to periodically execute its script.
func (cm *CollectorManager) RegisterAll() {
	for clusterName, clusterCfg := range cm.Config.Clusters {
		if !clusterCfg.Enabled {
			log.Printf("Cluster %s is disabled, skipping...", clusterName)
			continue
		}
		for collectorName, collectorCfg := range clusterCfg.Collectors {
			if !collectorCfg.Enabled {
				log.Printf("Collector %s in cluster %s is disabled, skipping...", collectorName, clusterName)
				continue
			}
			go cm.runCollector(clusterName, collectorName, collectorCfg)
		}
	}
}

// runCollector starts a ticker to periodically execute the collector's script.
func (cm *CollectorManager) runCollector(clusterName, collectorName string, collectorCfg config.CollectorConfig) {
	ticker := time.NewTicker(time.Duration(collectorCfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting collector %s in cluster %s with interval %ds", collectorName, clusterName, collectorCfg.Interval)
	key := fmt.Sprintf("%s:%s", clusterName, collectorName)
	for range ticker.C {
		output, execTime, err := cm.ScriptExecutor.ExecuteScript(collectorCfg.ScriptPath, collectorCfg.ScriptType, collectorCfg.Timeout)
		if err != nil {
			log.Printf("Error executing script %s for collector %s in cluster %s: %v", collectorCfg.ScriptPath, collectorName, clusterName, err)
			continue
		}

		// Add HELP and TYPE information for Prometheus metrics
		helpText := fmt.Sprintf(`# HELP %s Metric collected from external script`, collectorName)
		typeText := fmt.Sprintf(`# TYPE %s gauge`, collectorName)

		// Format the output with HELP, TYPE, script path and execution timestamp.
		formatted := fmt.Sprintf("%s\n%s\n# Script: %s, exec_time: %s\n%s", helpText, typeText, collectorCfg.ScriptPath, execTime, output)

		CollectorOutputs.Store(key, formatted)
		log.Printf("Updated output for %s", key)
	}
}

