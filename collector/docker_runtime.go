package collector

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// DockerRuntimeCollector 采集 Docker Runtime 信息
type DockerRuntimeCollector struct {
	runtime *prometheus.Desc
}

// NewDockerRuntimeCollector 创建新的 DockerRuntime 采集器
func NewDockerRuntimeCollector() *DockerRuntimeCollector {
	return &DockerRuntimeCollector{
		runtime: prometheus.NewDesc(
			"docker_default_runtime",
			"Default Docker Runtime",
			[]string{"runtime"}, nil,
		),
	}
}

// Describe 实现 Prometheus 采集器接口
func (c *DockerRuntimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.runtime
}

// Collect 采集数据
func (c *DockerRuntimeCollector) Collect(ch chan<- prometheus.Metric) {
	runtime, err := getDockerRuntime()
	if err != nil {
		log.Errorf("Failed to get Docker runtime: %v", err)
		return
	}

	// 暴露指标，值设为 1（仅用于标记存在）
	ch <- prometheus.MustNewConstMetric(c.runtime, prometheus.GaugeValue, 1, runtime)
}

// getDockerRuntime 获取默认 Docker Runtime
func getDockerRuntime() (string, error) {
	cmd := exec.Command("docker", "info")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// 解析 `Default Runtime` 行
	re := regexp.MustCompile(`Default Runtime:\s+(\S+)`)
	matches := re.FindStringSubmatch(out.String())
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}
	return "unknown", nil
}