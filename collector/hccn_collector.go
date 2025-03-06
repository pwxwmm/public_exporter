package collector

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// HCCNCollector 结构体
type HCCNCollector struct {
	deviceID int
	linkUp   *prometheus.Desc
	linkDown *prometheus.Desc
	linkDiff *prometheus.Desc
}

// NewHCCNCollector 创建新的 HCCN 采集器
func NewHCCNCollector(deviceID int) *HCCNCollector {
	return &HCCNCollector{
		deviceID: deviceID,
		linkUp: prometheus.NewDesc(
			"hccn_link_up_total",
			"Number of times the link was up",
			[]string{"device"}, nil,
		),
		linkDown: prometheus.NewDesc(
			"hccn_link_down_total",
			"Number of times the link was down",
			[]string{"device"}, nil,
		),
		linkDiff: prometheus.NewDesc(
			"hccn_link_change_duration_seconds",
			"Time difference between last two link status changes",
			[]string{"device"}, nil,
		),
	}
}

// Describe 实现 Prometheus 采集器接口
func (c *HCCNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.linkUp
	ch <- c.linkDown
	ch <- c.linkDiff
}

// Collect 采集数据
func (c *HCCNCollector) Collect(ch chan<- prometheus.Metric) {
	output, err := runHccnTool(c.deviceID)
	if err != nil {
		log.Errorf("Failed to run hccn_tool: %v", err)
		return
	}

	upCount, downCount, timeDiff := parseHccnOutput(output)
	device := fmt.Sprintf("%d", c.deviceID)

	ch <- prometheus.MustNewConstMetric(c.linkUp, prometheus.GaugeValue, float64(upCount), device)
	ch <- prometheus.MustNewConstMetric(c.linkDown, prometheus.GaugeValue, float64(downCount), device)

	if timeDiff >= 0 {
		ch <- prometheus.MustNewConstMetric(c.linkDiff, prometheus.GaugeValue, float64(timeDiff), device)
	}
}

// runHccnTool 执行 hccn_tool 命令
func runHccnTool(deviceID int) (string, error) {
	cmd := exec.Command("hccn_tool", "-i", fmt.Sprintf("%d", deviceID), "-link_stat", "-g")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing hccn_tool: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// parseHccnOutput 解析 hccn_tool 输出
func parseHccnOutput(output string) (int, int, int64) {
	upCount, downCount := 0, 0
	var timestamps []time.Time

	scanner := bufio.NewScanner(strings.NewReader(output))
	reUp := regexp.MustCompile(`link up count\s+:\s+(\d+)`)
	reDown := regexp.MustCompile(`link down count\s+:\s+(\d+)`)
	reTime := regexp.MustCompile(`$begin:math:display$(.+?)$end:math:display$\s+(LINK UP|LINK DOWN)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 解析 up/down count
		if match := reUp.FindStringSubmatch(line); match != nil {
			upCount, _ = strconv.Atoi(match[1])
		}
		if match := reDown.FindStringSubmatch(line); match != nil {
			downCount, _ = strconv.Atoi(match[1])
		}

		// 解析变更时间
		if match := reTime.FindStringSubmatch(line); match != nil {
			parsedTime, err := time.Parse("Mon Jan 2 15:04:05 2006", match[1])
			if err == nil {
				timestamps = append(timestamps, parsedTime)
			}
		}
	}

	// 计算最近两次变更时间差
	timeDiff := int64(-1)
	if len(timestamps) >= 2 {
		timeDiff = int64(timestamps[len(timestamps)-1].Sub(timestamps[len(timestamps)-2]).Seconds())
	}

	return upCount, downCount, timeDiff
}