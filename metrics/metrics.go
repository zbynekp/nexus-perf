package metrics

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// OperationType represents the type of operation
type OperationType string

const (
	OperationUpload   OperationType = "UPLOAD"
	OperationDownload OperationType = "DOWNLOAD"
)

// ResultMetrics holds metrics for a single operation result
type ResultMetrics struct {
	Duration       time.Duration
	BytesCount     int64
	StatusCode     int
	Success        bool
	ThroughputMbps float64
}

// AggregatedMetrics holds aggregated metrics for all operations
type AggregatedMetrics struct {
	OperationType   OperationType
	TotalOperations int
	SuccessCount    int
	FailureCount    int
	TotalBytes      int64
	TotalDuration   time.Duration

	MinThroughput float64
	MaxThroughput float64
	AvgThroughput float64

	MinDuration time.Duration
	MaxDuration time.Duration
	AvgDuration time.Duration

	P50Duration time.Duration
	P95Duration time.Duration
	P99Duration time.Duration

	OverallThroughputMbps float64
	SuccessRate           float64
}

// Collector collects and aggregates metrics
type Collector struct {
	mu      sync.RWMutex
	metrics []ResultMetrics
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		metrics: make([]ResultMetrics, 0),
	}
}

// Record records a single operation result
func (c *Collector) Record(duration time.Duration, bytes int64, success bool, statusCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	throughput := calculateThroughput(bytes, duration)
	c.metrics = append(c.metrics, ResultMetrics{
		Duration:       duration,
		BytesCount:     bytes,
		StatusCode:     statusCode,
		Success:        success,
		ThroughputMbps: throughput,
	})
}

// GetAggregated returns aggregated metrics
func (c *Collector) GetAggregated(opType OperationType) *AggregatedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.metrics) == 0 {
		return &AggregatedMetrics{
			OperationType: opType,
		}
	}

	// Calculate aggregates
	totalBytes := int64(0)
	totalDuration := time.Duration(0)
	successCount := 0
	durations := make([]time.Duration, 0, len(c.metrics))
	throughputs := make([]float64, 0, len(c.metrics))

	for _, m := range c.metrics {
		if m.Success {
			successCount++
		}
		totalBytes += m.BytesCount
		totalDuration += m.Duration
		durations = append(durations, m.Duration)
		throughputs = append(throughputs, m.ThroughputMbps)
	}

	// Sort for percentile calculations
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	sort.Float64s(throughputs)

	return &AggregatedMetrics{
		OperationType:         opType,
		TotalOperations:       len(c.metrics),
		SuccessCount:          successCount,
		FailureCount:          len(c.metrics) - successCount,
		TotalBytes:            totalBytes,
		TotalDuration:         totalDuration,
		MinThroughput:         throughputs[0],
		MaxThroughput:         throughputs[len(throughputs)-1],
		AvgThroughput:         calculateAverage(throughputs),
		MinDuration:           durations[0],
		MaxDuration:           durations[len(durations)-1],
		AvgDuration:           calculateAvgDuration(durations),
		P50Duration:           percentileDuration(durations, 50),
		P95Duration:           percentileDuration(durations, 95),
		P99Duration:           percentileDuration(durations, 99),
		OverallThroughputMbps: calculateThroughput(totalBytes, totalDuration),
		SuccessRate:           float64(successCount) / float64(len(c.metrics)) * 100,
	}
}

// calculateThroughput calculates throughput in Mbps
func calculateThroughput(bytes int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}

	// Convert bytes to megabytes
	megabytes := float64(bytes) / (1024 * 1024)

	// Convert duration to seconds
	seconds := duration.Seconds()

	// Mbps = megabytes / seconds
	return megabytes / seconds
}

// calculateAverage calculates average throughput
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// calculateAvgDuration calculates average duration
func calculateAvgDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	sum := int64(0)
	for _, d := range durations {
		sum += int64(d)
	}

	return time.Duration(sum / int64(len(durations)))
}

// percentileDuration calculates percentile duration
func percentileDuration(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	index := int(float64(len(durations)) * percentile / 100)
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

// String returns a formatted string representation of aggregated metrics
func (m *AggregatedMetrics) String() string {
	return fmt.Sprintf(`
=== %s Metrics ===
Total Operations:       %d
Success:                %d (%.2f%%)
Failures:               %d
Total Data:             %.2f MB
Total Duration:         %v

Throughput:
  Overall:              %.2f Mbps
  Average:              %.2f Mbps
  Min:                  %.2f Mbps
  Max:                  %.2f Mbps

Duration:
  Average:              %v
  Min:                  %v
  Max:                  %v
  P50:                  %v
  P95:                  %v
  P99:                  %v
`, m.OperationType, m.TotalOperations, m.SuccessCount, m.SuccessRate, m.FailureCount,
		float64(m.TotalBytes)/(1024*1024), m.TotalDuration,
		m.OverallThroughputMbps, m.AvgThroughput, m.MinThroughput, m.MaxThroughput,
		m.AvgDuration, m.MinDuration, m.MaxDuration, m.P50Duration, m.P95Duration, m.P99Duration)
}
