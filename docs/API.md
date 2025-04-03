
---

# Public Exporter - API Documentation

## Overview

`public_exporter` is a service that allows you to run custom scripts (shell or Python) to collect metrics from different sources, process that data, and expose it in a format that Prometheus can scrape. This document describes the API endpoints provided by the `public_exporter`.

## API Endpoints

### 1. **GET /metrics**

The `/metrics` endpoint is the primary API endpoint exposed by the `public_exporter`. It returns the metrics collected from the custom scripts in **Prometheus format**.

#### Endpoint:
```
GET /metrics
```

#### Response:
The response will be a plain text document containing Prometheus-formatted metric data. For example:

```txt
# HELP cluster_1_demo_collector_temperature Metric collected from external script (script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123")
# TYPE cluster_1_demo_collector_temperature gauge
cluster_1_demo_collector_temperature{gpu="0", script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123"} 55
cluster_1_demo_collector_temperature{gpu="1",type="test",testkey1="testvalue1", script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123"} 60
cluster_1_demo_collector_temperature{script_path="/opt/scripts/demo_collector.sh", exec_time="2025-04-01 14:35:21.123"} 65
```

#### Description:
- **Metric Name**: Each metric is named according to the configuration in the `config.yaml` file.
- **Labels**: Metrics can have labels, which provide additional context to the metric. Labels are key-value pairs, such as `gpu="0"`, `type="test"`, etc.
- **Metric Value**: The numeric value of the metric (e.g., `55`, `60`, `65`).
- **`script_path`**: The path of the script that collected this metric.
- **`exec_time`**: The timestamp when the script was executed, in the format `YYYY-MM-DD HH:MM:SS.MMM`.

#### Use Case:
This endpoint is typically scraped by Prometheus at the interval specified in the `config.yaml` file. The data will be collected and exposed in Prometheus format, allowing you to monitor and visualize your system's health and performance.

---

### 2. **Health Check (Optional)**

To check if the `public_exporter` is running and healthy, you can implement a basic health check endpoint. This is useful to ensure that the exporter is up and running before integrating it into a larger monitoring system.

#### Endpoint:
```
GET /health
```

#### Response:
```json
{
  "status": "ok"
}
```

#### Description:
- The `/health` endpoint returns a simple JSON object with a `status` key indicating whether the exporter is healthy and operational.
- This can be used by monitoring systems or orchestration tools like Kubernetes to verify the exporter’s health.

---

## Configuration File

The `config.yaml` file is used to configure the behavior of `public_exporter`. The file defines which clusters and collectors are enabled, the paths to the scripts, the interval at which they are executed, and more.

### Configuration Format

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

### Configuration Fields:

- **`global`**: Settings that apply to all clusters:
  - `log_level`: The logging level (`debug`, `info`, `warning`, `error`).
  - `log_file`: The path to the log file.
  - `default_scrape_interval`: Default collection interval in seconds (used if a specific interval is not specified for a collector).

- **`clusters`**: Each cluster has its own configuration:
  - `enabled`: Whether the cluster is enabled or not.
  - **Collectors**: These are the various data collectors configured for the cluster. Each collector has the following fields:
    - `enabled`: Whether the collector is enabled or not.
    - `interval`: The interval (in seconds) at which the script should be executed.
    - `timeout`: The maximum time (in seconds) the script is allowed to run before being terminated.
    - `script_path`: The path to the script that will be executed.
    - `script_type`: The type of the script (`shell` or `python`).

---

## Error Handling

In case of an error (e.g., script execution fails, timeout occurs), the exporter logs the error and returns the following Prometheus-formatted response for the metric:

```txt
# HELP <metric_name> Error occurred while collecting metric from script
# TYPE <metric_name> gauge
<metric_name>{error="script execution failed", script_path="<path_to_script>", exec_time="<timestamp>"} 0
```

This allows users to identify when an error occurs during data collection and take appropriate action.

---

## Conclusion

The `public_exporter` API provides a simple and flexible way to collect custom metrics from scripts and expose them for Prometheus scraping. You can define multiple clusters and collectors, with specific intervals and timeouts for each script, and easily monitor system health.

If you have any questions or need further assistance, feel free to reach out to the author at `mmwei3@iflytek.com`.