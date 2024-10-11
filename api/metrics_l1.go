package api

import (
	"context"

	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cpuGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_cores",
		Help: "Number of CPU cores",
	}, []string{"nodeID"})

	cpuUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_usage",
		Help: "CPU usage in percentage",
	}, []string{"nodeID"})

	memoryGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "memory",
		Help: "Number of Memory",
	}, []string{"nodeID"})

	memoryUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "memory_usage",
		Help: "Memory usage in percentage",
	}, []string{"nodeID"})

	diskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "disk",
		Help: "Number of Disk",
	}, []string{"nodeID"})

	diskUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "disk_usage",
		Help: "Disk usage in percentage",
	}, []string{"nodeID"})
)

func init() {
	prometheus.MustRegister(cpuGauge)
	prometheus.MustRegister(cpuUsageGauge)
	prometheus.MustRegister(memoryGauge)
	prometheus.MustRegister(memoryUsageGauge)
	prometheus.MustRegister(diskGauge)
	prometheus.MustRegister(diskUsageGauge)
}

func setL1Gatherer(ctx context.Context) {
	for _, node := range statistics.AllNodesMap {
		cpuGauge.WithLabelValues(node.DeviceID).Set(float64(node.CpuCores))
		cpuUsageGauge.WithLabelValues(node.DeviceID).Set(node.CpuUsage)
		memoryGauge.WithLabelValues(node.DeviceID).Set(node.Memory)
		memoryUsageGauge.WithLabelValues(node.DeviceID).Set(node.MemoryUsage)
		diskGauge.WithLabelValues(node.DeviceID).Set(node.DiskSpace)
		diskUsageGauge.WithLabelValues(node.DeviceID).Set(node.DiskUsage)
	}
}
