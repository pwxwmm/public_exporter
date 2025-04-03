
---

# Public Exporter - Custom Script Integration Guide

## Author Information

- **Author**: mmwei3
- **Email**: mmwei3@iflytek.com, 1300042631@qq.com
- **Date**: 2025-03-28

## Overview

The `public_exporter` allows you to integrate custom scripts for data collection. These scripts are executed periodically, and the data they produce is exposed in the **Prometheus** format for monitoring. This guide provides the format your script's output should follow, and how the data will be exposed by the exporter.

## Script Output Format

Your script should produce output in the following format:

1. **Metric Name**
2. **Labels (optional)**
3. **Metric Value**

The basic syntax for each output line is:

```
<metric_name>{<label1>="<value1>", <label2>="<value2>", ...} <metric_value>
```

### Examples

#### 1. Metric with Labels
```bash
cluster_1_demo_collector_temperature{gpu="0"} 55
```
- **Metric Name**: `cluster_1_demo_collector_temperature`
- **Label**: `gpu="0"`
- **Metric Value**: `55`

#### 2. Metric with Multiple Labels
```bash
cluster_1_demo_collector_temperature{gpu="1", type="test", testkey1="testvalue1"} 60
```
- **Metric Name**: `cluster_1_demo_collector_temperature`
- **Labels**: `gpu="1"`, `type="test"`, `testkey1="testvalue1"`
- **Metric Value**: `60`

#### 3. Metric without Labels
```bash
cluster_1_demo_collector_temperature 65
```
- **Metric Name**: `cluster_1_demo_collector_temperature`
- **Metric Value**: `65`

## Script Execution and Output Handling

### Supported Script Types
You can use the following script types:
- **Shell Scripts** (`.sh`)
- **Python Scripts** (`.py`)

Ensure the output is consistent with the format shown above, with one line per metric.

### Example Shell Script (`/opt/scripts/demo_collector.sh`):
```bash
#!/bin/bash

echo 'cluster_1_demo_collector_temperature{gpu="0"} 55'
echo 'cluster_1_demo_collector_temperature{gpu="1",type="test",testkey1="testvalue1"} 60'
echo 'cluster_1_demo_collector_temperature 65'
```

### Example Python Script (`/opt/scripts/demo_collector.py`):
```python
#!/usr/bin/env python3

print('cluster_1_demo_collector_temperature{gpu="0"} 55')
print('cluster_1_demo_collector_temperature{gpu="1",type="test",testkey1="testvalue1"} 60')
print('cluster_1_demo_collector_temperature 65')
```

## Exporter Data Exposure

### Prometheus Format

When the script is executed, the exporter will expose the data in the following format at the `/metrics` endpoint:

#### Example Output (from `/metrics` endpoint):

```
# HELP cluster_1_demo_collector_temperature Metric collected from external script 
# (script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123")
# TYPE cluster_1_demo_collector_temperature gauge
cluster_1_demo_collector_temperature{gpu="0"} 55
cluster_1_demo_collector_temperature{gpu="1",type="test",testkey1="testvalue1"} 60
cluster_1_demo_collector_temperature{} 65
```

### Explanation of Output Fields:

- **Metric Name**: The name of the metric being reported (e.g., `cluster_1_demo_collector_temperature`).
- **Labels**: Labels are key-value pairs that provide additional context to the metric (e.g., `gpu="0"`, `type="test"`).
- **Metric Value**: The value of the metric (e.g., `55`, `60`, `65`).
- **`script_path`**: The path to the script that generated this metric.
- **`exec_time`**: The timestamp when the script was executed, in the format `YYYY-MM-DD HH:MM:SS.MMM`.

### Metric Naming Convention

- **Metric names** should be in lowercase and can include underscores. For example, `cluster_1_demo_collector_temperature`.
- **Labels** should use key-value pairs in the format `key="value"`. The keys should also follow the lowercase convention.

### Multiple Metrics from One Script

A single script can output multiple metrics, and each metric will be exposed separately. For example, one script may report GPU temperature, CPU load, and memory usage, with each reported as a different metric.

Example output:
```
# HELP cluster_1_demo_collector_temperature Metric collected from external script (script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123")
# TYPE cluster_1_demo_collector_temperature gauge
cluster_1_demo_collector_temperature{gpu="0"} 55
cluster_1_demo_collector_temperature{gpu="1"} 60

# HELP cluster_1_demo_collector_cpu_usage Metric collected from external script (script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123")
# TYPE cluster_1_demo_collector_cpu_usage gauge
cluster_1_demo_collector_cpu_usage{core="0"} 80
cluster_1_demo_collector_cpu_usage{core="1"} 85
```

## Scraping Interval

The `public_exporter` will scrape data from your script based on the configuration defined in the `config.yaml` file. The interval defines how often the script will be executed.

### Example `config.yaml` Configuration

```yaml
# Global settings (apply to all clusters)
global:
  log_level: "info"  # Log level: debug, info, warning, error
  log_file: "/var/log/public_exporter.log"  # Log file path
  default_scrape_interval: 30  # Default collection interval in seconds (if not specified)

# Cluster-specific settings
clusters:
  cluster_A:
    enabled: true  # Enable or disable this cluster
    collectors:
      npu:
        enabled: true  # Enable or disable this collector
        interval: 30  # Execution interval (seconds)
        timeout: 10  # Execution timeout (seconds)
        script_path: "/usr/local/bin/npu_status.sh"
        script_type: "shell"  # Script type: shell or python

      gpu:
        enabled: true
        interval: 60  # Execution interval (seconds)
        timeout: 15  # Execution timeout (seconds)
        script_path: "/usr/local/bin/gpu_status.sh"
        script_type: "shell"

      temp:
        enabled: true
        script_path: "/opt/scripts/temp_collector.py"  # Path to Python script
        script_type: "python"  # Script type: python
        interval: 30  # Execution interval (seconds)
        timeout: 10  # Execution timeout (seconds)

  cluster_B:
    enabled: false  # This cluster is disabled, no data collection
    collectors:
      gpu:
        enabled: true
        interval: 45  # Execution interval (seconds)
        timeout: 10  # Execution timeout (seconds)
        script_path: "/usr/local/bin/gpu_status_cluster_B.sh"
        script_type: "shell"
      cpu:
        enabled: true
        interval: 20  # Execution interval (seconds)
        script_type: "shell"
        script_path: "/usr/local/bin/cpu_status.sh"
```

### Configuration Explanation

- **`enabled`**: Set to `true` or `false` to enable or disable a specific cluster or collector.
- **`script_type`**: Specify whether the script is a `shell` or `python` script.
- **`interval`**: Defines the interval in seconds at which the script will be executed.
- **`timeout`**: The maximum time (in seconds) allowed for the script to run before being terminated.
- **`script_path`**: The path to the script that will be executed.

### Scraping and Timeout

- **Interval**: Defines how often the script should be executed. For example, if you set `interval: 30`, the script will run every 30 seconds.
- **Timeout**: Defines the maximum time the script can run before it is forcibly terminated. This is set in the `config.yaml` file.

### Example:

```yaml
interval: 30
timeout: 10
script_type: shell  # or python
script_path: "/opt/scripts/demo_collector.sh"
```

In this example:
- The script `demo_collector.sh` will run every 30 seconds.
- If the script takes longer than 10 seconds, it will be forcibly terminated.


## Troubleshooting

If the script is not producing the expected output:
1. Make sure the script outputs data in the correct format (metric name, labels, and value).
2. Check the log file (`/var/log/public_exporter.log`) for any errors related to script execution.
3. Ensure the script is executable and accessible by the `public_exporter` process.

## Conclusion

By following the above guidelines, you can create custom scripts for your monitoring needs and integrate them seamlessly into the `public_exporter` to expose data for Prometheus.