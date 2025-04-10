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

var (
	CollectorOutputs sync.Map // key: "cluster:collector" -> formatted output
	CollectorHealth  sync.Map // key: "cluster:collector" -> 1 or 0
)

type ScriptExecutor struct {
	mu sync.Mutex
}

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

	return string(output), execTime, nil
}

type CollectorManager struct {
	Config         *config.Config
	ScriptExecutor *ScriptExecutor
}

func NewCollectorManager(cfg *config.Config) *CollectorManager {
	return &CollectorManager{
		Config:         cfg,
		ScriptExecutor: &ScriptExecutor{},
	}
}

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

func (cm *CollectorManager) runCollector(clusterName, collectorName string, collectorCfg config.CollectorConfig) {
	ticker := time.NewTicker(time.Duration(collectorCfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting collector %s in cluster %s with interval %ds", collectorName, clusterName, collectorCfg.Interval)
	key := fmt.Sprintf("%s:%s", clusterName, collectorName)

	for range ticker.C {
		output, execTime, err := cm.ScriptExecutor.ExecuteScript(collectorCfg.ScriptPath, collectorCfg.ScriptType, collectorCfg.Timeout)
		if err != nil {
			log.Printf("Error executing script %s for collector %s in cluster %s: %v", collectorCfg.ScriptPath, collectorName, clusterName, err)
			CollectorHealth.Store(key, 0)
			continue
		}

		helpText := fmt.Sprintf(`# HELP %s Metric collected from external script`, collectorName)
		typeText := fmt.Sprintf(`# TYPE %s gauge`, collectorName)
		formatted := fmt.Sprintf("%s\n%s\n# Script: %s, exec_time: %s\n%s", helpText, typeText, collectorCfg.ScriptPath, execTime, output)

		CollectorOutputs.Store(key, formatted)
		CollectorHealth.Store(key, 1)
		log.Printf("Updated output for %s", key)
	}
}