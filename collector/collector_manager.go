// Author: mmwei3
// Email: mmwei3@iflytek.com
// Date: 2025-04-03
//
// Description:
// This package manages data collectors that execute external scripts periodically. 
// It supports Python and Shell scripts, ensuring their outputs are stored and exposed as metrics.
// The collected data is formatted for Prometheus compatibility and stored in a thread-safe manner.

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

// CollectorOutput represents the output of a collector execution
type CollectorOutput struct {
	Output   string
	ExecTime string
	LastSeen time.Time
	Error    error
}

// CollectorManager manages all data collectors
type CollectorManager struct {
	Config         *config.Config
	ScriptExecutor *ScriptExecutor
	outputs        sync.Map // key: "cluster:collector" -> *CollectorOutput
	health         sync.Map // key: "cluster:collector" -> int (1 or 0)
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

// ScriptExecutor handles script execution with proper timeout and error handling
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
	switch scriptType {
	case "python", "python3":
		cmd = exec.CommandContext(ctx, "python3", scriptPath)
	case "python2":
		cmd = exec.CommandContext(ctx, "python2", scriptPath)
	case "shell":
		cmd = exec.CommandContext(ctx, "bash", scriptPath)
	default:
		return "", "", fmt.Errorf("unsupported script type: %s", scriptType)
	}

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", "", fmt.Errorf("script execution timed out after %d seconds", timeout)
	}
	if err != nil {
		return "", "", fmt.Errorf("script execution failed: %v, output: %s", err, string(output))
	}

	return string(output), execTime, nil
}

func NewCollectorManager(cfg *config.Config) *CollectorManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &CollectorManager{
		Config:         cfg,
		ScriptExecutor: &ScriptExecutor{},
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts all enabled collectors
func (cm *CollectorManager) Start() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

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
			
			// Validate collector configuration
			if err := cm.validateCollectorConfig(collectorCfg); err != nil {
				log.Printf("Invalid configuration for collector %s in cluster %s: %v", collectorName, clusterName, err)
				continue
			}
			
			cm.wg.Add(1)
			go cm.runCollector(clusterName, collectorName, collectorCfg)
		}
	}
	
	return nil
}

// Stop gracefully stops all collectors
func (cm *CollectorManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	log.Println("Stopping all collectors...")
	cm.cancel()
	cm.wg.Wait()
	log.Println("All collectors stopped")
}

// validateCollectorConfig validates collector configuration
func (cm *CollectorManager) validateCollectorConfig(cfg config.CollectorConfig) error {
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
	return nil
}

func (cm *CollectorManager) runCollector(clusterName, collectorName string, collectorCfg config.CollectorConfig) {
	defer cm.wg.Done()
	
	ticker := time.NewTicker(time.Duration(collectorCfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting collector %s in cluster %s with interval %ds", collectorName, clusterName, collectorCfg.Interval)
	key := fmt.Sprintf("%s:%s", clusterName, collectorName)

	// Execute once immediately
	cm.executeCollector(key, collectorName, collectorCfg)

	for {
		select {
		case <-ticker.C:
			cm.executeCollector(key, collectorName, collectorCfg)
		case <-cm.ctx.Done():
			log.Printf("Collector %s in cluster %s stopped", collectorName, clusterName)
			return
		}
	}
}

func (cm *CollectorManager) executeCollector(key, collectorName string, collectorCfg config.CollectorConfig) {
	output, execTime, err := cm.ScriptExecutor.ExecuteScript(collectorCfg.ScriptPath, collectorCfg.ScriptType, collectorCfg.Timeout)
	
	collectorOutput := &CollectorOutput{
		Output:   output,
		ExecTime: execTime,
		LastSeen: time.Now(),
		Error:    err,
	}
	
	if err != nil {
		log.Printf("Error executing script %s for collector %s: %v", collectorCfg.ScriptPath, collectorName, err)
		cm.health.Store(key, 0)
		collectorOutput.Output = fmt.Sprintf("Error: %v", err)
	} else {
		// Format for Prometheus
		helpText := fmt.Sprintf(`# HELP %s Metric collected from external script`, collectorName)
		typeText := fmt.Sprintf(`# TYPE %s gauge`, collectorName)
		formatted := fmt.Sprintf("%s\n%s\n# Script: %s, exec_time: %s\n%s", 
			helpText, typeText, collectorCfg.ScriptPath, execTime, output)
		collectorOutput.Output = formatted
		cm.health.Store(key, 1)
	}
	
	cm.outputs.Store(key, collectorOutput)
	log.Printf("Updated output for %s", key)
}

// GetOutputs returns all collector outputs for metrics endpoint
func (cm *CollectorManager) GetOutputs() []string {
	var outputs []string
	cm.outputs.Range(func(key, value interface{}) bool {
		if output, ok := value.(*CollectorOutput); ok {
			outputs = append(outputs, output.Output)
		}
		return true
	})
	return outputs
}

// GetHealthStatus returns health status for all collectors
func (cm *CollectorManager) GetHealthStatus() map[string]int {
	status := make(map[string]int)
	cm.health.Range(func(key, value interface{}) bool {
		if health, ok := value.(int); ok {
			status[key.(string)] = health
		}
		return true
	})
	return status
}

// GetCollectorCount returns the total number of active collectors
func (cm *CollectorManager) GetCollectorCount() int {
	count := 0
	cm.health.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}