# Changelog

## [Unreleased]

### Added
- Prometheus 指标 `collector_health_status{cluster, collector}`，用于标识每个 collector 的健康状态。
- Prometheus 指标 `exporter_health_status`，用于表示 exporter 全局健康状态（任一 collector 异常为 0）。
- `/health` 接口支持返回每个 collector 的健康状态 JSON 结构。

### Changed
- 优化 `collector_manager.go` 中 collector 输出逻辑，确保采集稳定。
- 更新 `main.go`，在 `/metrics` 中加入 exporter 全局健康状态 `exporter_health_status`。

### Demo info
```python
[root@k8s-master01 public_exporter]# docker logs -f  prometheus_public_exporter
=====================================
         Public Exporter
=====================================
Author: mmwei3
Email: mmwei3@iflytek.com, 1300042631@qq.com
Date: 2025-03-28
2025/04/10 11:11:04 Starting public_exporter...
2025/04/10 11:11:04 Exporter service started.
2025/04/10 11:11:04 Cluster cluster_B is disabled, skipping...
2025/04/10 11:11:04 Exporter is running on http://0.0.0.0:5535/metrics
2025/04/10 11:11:04 Cluster cluster_B is disabled, skipping...
2025/04/10 11:11:04 Starting collector gpu in cluster cluster_A with interval 68s
2025/04/10 11:11:04 Scheduler: starting collector npu in cluster cluster_A with interval 37s
2025/04/10 11:11:04 Starting collector temp in cluster cluster_A with interval 56s
2025/04/10 11:11:04 Scheduler: starting collector temp in cluster cluster_A with interval 56s
2025/04/10 11:11:04 Starting collector npu in cluster cluster_A with interval 37s
2025/04/10 11:11:04 Scheduler: starting collector gpu in cluster cluster_A with interval 68s



2025/04/10 11:11:41 Updated output for cluster_A:npu
2025/04/10 11:11:41 Scheduler: updated output for cluster_A:npu
2025/04/10 11:11:41 Collector npu executed at 2025-04-10 11:11:41.407959011 +0800 CST m=+37.008759229, next execution in 37s


2025/04/10 11:12:00 Scheduler: updated output for cluster_A:temp
2025/04/10 11:12:00 Updated output for cluster_A:temp
2025/04/10 11:12:00 Collector temp executed at 2025-04-10 11:12:00.429425447 +0800 CST m=+56.030225670, next execution in 56s
2025/04/10 11:12:12 Scheduler: updated output for cluster_A:gpu
2025/04/10 11:12:12 Collector gpu executed at 2025-04-10 11:12:12.409157652 +0800 CST m=+68.009957878, next execution in 68s
2025/04/10 11:12:12 Updated output for cluster_A:gpu
2025/04/10 11:12:18 Scheduler: updated output for cluster_A:npu
2025/04/10 11:12:18 Collector npu executed at 2025-04-10 11:12:18.407401842 +0800 CST m=+74.008202063, next execution in 37s
2025/04/10 11:12:18 Updated output for cluster_A:npu
^C
[root@k8s-master01 public_exporter]# curl http://172.29.228.139:5535/health
{"status":"ok", "collectors":{"cluster_A:npu":"ok","cluster_A:gpu":"ok","cluster_A:temp":"ok"}}[root@k8s-master01 public_exporter]#
[root@k8s-master01 public_exporter]#
[root@k8s-master01 public_exporter]# curl http://172.29.228.139:5535/metrics
# HELP gpu Metric collected from external script
# TYPE gpu gauge
# Script: /opt/scripts/shell/gpu_status.sh, exec_time: 2025-04-10 11:12:12.406
cluster_1_demo_gpu_collector_temperature{gpu="0"} 55
cluster_1_demo_gpu_collector_temperature{gpu="1",type="test",testkey1="testvalue1"} 60
cluster_1_demo_gpu_collector_temperature 65

# HELP npu Metric collected from external script
# TYPE npu gauge
# Script: /opt/scripts/shell/npu_status.sh, exec_time: 2025-04-10 11:12:18.404
cluster_1_demo_npu_collector_temperature{npu="0"} 15
cluster_1_demo_npu_collector_temperature{npu="1",type="test",testkey1="testvalue1"} 20
cluster_1_demo_npu_collector_temperature 95

# HELP temp Metric collected from external script
# TYPE temp gauge
# Script: /opt/scripts/python/temp_collector.py, exec_time: 2025-04-10 11:12:00.403
cluster_1_demo_py_collector_temperature{gpu="0"} 55
cluster_1_demo_py_collector_temperature{gpu="1",type="test",testkey1="testvalue1"} 60
cluster_1_demo_py_collector_temperature 65

collector_health_status{cluster="cluster_A", collector="npu"} 1
collector_health_status{cluster="cluster_A", collector="gpu"} 1
collector_health_status{cluster="cluster_A", collector="temp"} 1
# HELP exporter_health_status Global health status of the exporter
# TYPE exporter_health_status gauge

```

## [1.0.0] - 2025-04-10

### Added
- 基本的 `/metrics` 和 `/health` 端点，支持 Prometheus 采集器数据和健康状态。
- 初始版本中，支持通过配置文件管理 collector。

### Changed
- 完成了 collector 和健康状态的改进，确保在采集失败时能反映在 Prometheus 指标中。

## [0.1.0] - 2025-03-28

### Added
- 项目初始化，包括基础的 `collector_manager.go` 和配置文件处理。
- 初步支持通过定时执行脚本采集 Prometheus 兼容的数据。

