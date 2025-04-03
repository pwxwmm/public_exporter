package scheduler

import (
	"fmt"
	"log"
	"public_exporter/config"
	"public_exporter/collector"
	"time"
)

// Scheduler is used to coordinate periodic tasks.
type Scheduler struct{}

// NewScheduler returns a new Scheduler instance.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// StartScheduler iterates over all enabled clusters and collectors,
// starting a goroutine for each collector to periodically execute its script.
func (s *Scheduler) StartScheduler(cfg *config.Config) {
	// Use the ScriptExecutor from the collector package.
	scriptExecutor := &collector.ScriptExecutor{}

	for clusterName, clusterCfg := range cfg.Clusters {
		if !clusterCfg.Enabled {
			log.Printf("Cluster %s is disabled, skipping...", clusterName)
			continue
		}
		for collectorName, collCfg := range clusterCfg.Collectors {
			if !collCfg.Enabled {
				log.Printf("Collector %s in cluster %s is disabled, skipping...", collectorName, clusterName)
				continue
			}

			go func(clusterName, collectorName string, coll config.CollectorConfig) {
				// Initialize a ticker with the collector's interval
				ticker := time.NewTicker(time.Duration(coll.Interval) * time.Second)
				defer ticker.Stop()

				// Log starting collector info
				log.Printf("Scheduler: starting collector %s in cluster %s with interval %ds", collectorName, clusterName, coll.Interval)
				key := fmt.Sprintf("%s:%s", clusterName, collectorName)

				// Infinite loop to run the collector on each tick
				for {
					select {
					case <-ticker.C:
						// Execute the script
						output, execTime, err := scriptExecutor.ExecuteScript(coll.ScriptPath, coll.ScriptType, coll.Timeout)
						if err != nil {
							log.Printf("Scheduler: error executing script %s for collector %s in cluster %s: %v", coll.ScriptPath, collectorName, clusterName, err)
							continue
						}

						// Format the output and store it
						formatted := fmt.Sprintf("# Script: %s, exec_time: %s\n%s", coll.ScriptPath, execTime, output)
						collector.CollectorOutputs.Store(key, formatted)
						log.Printf("Scheduler: updated output for %s", key)

						// Log time of execution and next scheduled execution
						log.Printf("Collector %s executed at %v, next execution in %ds", collectorName, time.Now(), coll.Interval)
					}
				}
			}(clusterName, collectorName, collCfg)
		}
	}
}
