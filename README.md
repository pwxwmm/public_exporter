
---

# Public Exporter

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/Version-1.0.0-orange.svg)](CHANGELOG.md)

A high-performance Prometheus exporter for executing external scripts and collecting metrics from multiple clusters.

## Features

- 🚀 **High Performance**: Efficient script execution with proper timeout handling
- 🔧 **Flexible Configuration**: Support for multiple clusters and collectors
- 📊 **Prometheus Compatible**: Native Prometheus metrics format
- 🐍 **Multi-Language Support**: Python2, Python3, and Shell scripts
- 🏥 **Health Monitoring**: Built-in health checks and status monitoring
- 🔄 **Graceful Shutdown**: Proper cleanup and resource management
- 📝 **Structured Logging**: Log rotation and configurable log levels
- 🐳 **Docker Ready**: Containerized deployment support

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP Server   │    │ CollectorManager │    │ ScriptExecutor  │
│   (Port 5535)   │◄──►│                  │◄──►│                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   /metrics      │    │   Collectors     │    │  Python/Shell   │
│   /health       │    │   (Goroutines)   │    │    Scripts      │
│   /             │    └──────────────────┘    └─────────────────┘
└─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- Python2/Python3 (for Python script collectors)
- Bash (for shell script collectors)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd public_exporter
   ```

2. **Install dependencies**
   ```bash
   make deps
   # or manually:
   go mod download
   go mod tidy
   ```

3. **Build the application**
   ```bash
   make build
   # or manually:
   go build -o build/public_exporter ./cmd
   ```

4. **Run the exporter**
   ```bash
   make run
   # or manually:
   ./build/public_exporter -config.file=./config/config.yaml
   ```

### Configuration

Create a `config.yaml` file:

```yaml
global:
  log_file: "/var/log/public_exporter/exporter.log"
  log_level: "info"
  log_max_age: 7
  log_rotation_time: 24
  http_port: 5535
  http_timeout: 30
  default_scrape_interval: 60

clusters:
  production:
    enabled: true
    collectors:
      system_metrics:
        enabled: true
        interval: 30
        timeout: 10
        script_path: "/scripts/check_system.py"
        script_type: "python3"
      
      network_status:
        enabled: true
        interval: 60
        timeout: 15
        script_path: "/scripts/check_network.sh"
        script_type: "shell"
```

### Script Requirements

Your collection scripts should output metrics in Prometheus format:

```python
#!/usr/bin/env python3
# Example Python collector script

import time
import random

# Simulate some metrics
cpu_usage = random.uniform(0, 100)
memory_usage = random.uniform(0, 100)

# Output in Prometheus format
print(f"# HELP cpu_usage_percent CPU usage percentage")
print(f"# TYPE cpu_usage_percent gauge")
print(f"cpu_usage_percent {cpu_usage}")

print(f"# HELP memory_usage_percent Memory usage percentage")
print(f"# TYPE memory_usage_percent gauge")
print(f"memory_usage_percent {memory_usage}")
```

```bash
#!/bin/bash
# Example Shell collector script

# Get disk usage
disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')

# Output in Prometheus format
echo "# HELP disk_usage_percent Disk usage percentage"
echo "# TYPE disk_usage_percent gauge"
echo "disk_usage_percent $disk_usage"
```

## API Endpoints

### `/metrics`
Prometheus metrics endpoint. Returns all collected metrics in Prometheus format.

### `/health`
Health check endpoint. Returns JSON status of all collectors.

```json
{
  "status": "ok",
  "collectors": {
    "production:system_metrics": "ok",
    "production:network_status": "ok"
  }
}
```

### `/`
Root endpoint with basic information and links to other endpoints.

## Metrics

The exporter provides several built-in metrics:

- `collector_health_status{cluster="name", collector="name"}` - Health status of each collector (1=healthy, 0=unhealthy)
- `exporter_health_status` - Global health status of the exporter
- `collector_count` - Total number of active collectors

## Docker Deployment

### Build Docker Image
```bash
make docker-build
```

### Run Container
```bash
make docker-run
```

### Stop Container
```bash
make docker-clean
```

## Development

### Project Structure
```
public_exporter/
├── cmd/                    # Main application entry point
├── collector/             # Data collection management
├── config/                # Configuration management
├── service/               # Service layer coordination
├── scripts/               # Example collection scripts
├── build/                 # Build artifacts
├── config.yaml            # Configuration file
├── Dockerfile             # Docker configuration
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── README.md              # This file
```

### Available Make Targets

```bash
make help                  # Show all available targets
make build                 # Build the application
make test                  # Run tests
make coverage              # Run tests with coverage
make lint                  # Run linter
make run                   # Build and run
make docker-build          # Build Docker image
make docker-run            # Run Docker container
make install               # Install to /usr/local/bin
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test ./collector/...
```

## Configuration Reference

### Global Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `log_file` | string | - | Path to log file (required) |
| `log_level` | string | "info" | Log level (debug, info, warn, error, fatal, panic) |
| `log_max_age` | int | 7 | Log retention in days |
| `log_rotation_time` | int | 24 | Log rotation interval in hours |
| `http_port` | int | 5535 | HTTP server port |
| `http_timeout` | int | 30 | HTTP request timeout in seconds |
| `default_scrape_interval` | int | 60 | Default collection interval in seconds |

### Collector Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | false | Whether the collector is enabled |
| `interval` | int | global default | Collection interval in seconds |
| `timeout` | int | 30 | Script execution timeout in seconds |
| `script_path` | string | - | Path to the collection script (required) |
| `script_type` | string | - | Script type: python, python2, python3, shell (required) |

## Troubleshooting

### Common Issues

1. **Script execution fails**
   - Check script permissions (`chmod +x /path/to/script`)
   - Verify script path in configuration
   - Check script syntax and dependencies

2. **Metrics not appearing**
   - Verify collector is enabled in configuration
   - Check script output format (should be Prometheus compatible)
   - Review logs for execution errors

3. **High resource usage**
   - Adjust collection intervals
   - Optimize script execution time
   - Review script resource consumption

### Log Analysis

Enable debug logging to troubleshoot issues:

```yaml
global:
  log_level: "debug"
```

### Health Checks

Use the `/health` endpoint to monitor collector status:

```bash
curl http://localhost:5535/health
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

- **mmwei3** - *Initial work* - [mmwei3@iflytek.com](mailto:mmwei3@iflytek.com)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed list of changes and version history.

## Support

For support and questions:
- Create an issue in the repository
- Contact: mmwei3@iflytek.com
- Documentation: [docs/](docs/)

