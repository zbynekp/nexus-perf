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
	ThroughputMBps float64
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

	OverallThroughputMBps float64
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
		ThroughputMBps: throughput,
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

	var (
		totalBytes    int64
		totalDuration time.Duration
		successCount  int
		durations     = make([]time.Duration, 0, len(c.metrics))
		throughputs   = make([]float64, 0, len(c.metrics))
	)

	for _, m := range c.metrics {
		if m.Success {
			successCount++
		}
		totalBytes += m.BytesCount
		totalDuration += m.Duration
		durations = append(durations, m.Duration)
		throughputs = append(throughputs, m.ThroughputMBps)
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
		OverallThroughputMBps: calculateThroughput(totalBytes, totalDuration),
		SuccessRate:           float64(successCount) / float64(len(c.metrics)) * 100,
	}
}

// formatDuration formats a duration to 2 decimal places with the appropriate unit.
func formatDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.2fs", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	case d >= time.Microsecond:
		return fmt.Sprintf("%.2fµs", float64(d)/float64(time.Microsecond))
	default:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
}

// calculateThroughput calculates throughput in MB/s (SI megabytes per second)
func calculateThroughput(bytes int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(bytes) / 1_000_000 / duration.Seconds()
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
Total Duration:         %s

Throughput:
  Overall:              %.2f MB/s
  Average:              %.2f MB/s
  Min:                  %.2f MB/s
  Max:                  %.2f MB/s

Duration:
  Average:              %s
  Min:                  %s
  Max:                  %s
  P50:                  %s
  P95:                  %s
  P99:                  %s
`, m.OperationType, m.TotalOperations, m.SuccessCount, m.SuccessRate, m.FailureCount,
		float64(m.TotalBytes)/1_000_000, formatDuration(m.TotalDuration),
		m.OverallThroughputMBps, m.AvgThroughput, m.MinThroughput, m.MaxThroughput,
		formatDuration(m.AvgDuration), formatDuration(m.MinDuration), formatDuration(m.MaxDuration),
		formatDuration(m.P50Duration), formatDuration(m.P95Duration), formatDuration(m.P99Duration))
}

// FormatAsTable returns a human-readable table of results
func (m *AggregatedMetrics) FormatAsTable() string {
	return fmt.Sprintf(`
================================================================================
                      %s PERFORMANCE METRICS
================================================================================

OPERATIONS
--------------------------------------------------------------------------------
  Total Operations:           %d
  Success:                    %d (%.2f%%)
  Failures:                   %d
  Total Data Volume:          %.2f MB

THROUGHPUT (MB/s)
--------------------------------------------------------------------------------
  Overall:                    %.2f MB/s
  Average:                    %.2f MB/s
  Min:                        %.2f MB/s
  Max:                        %.2f MB/s

DURATION / LATENCY
--------------------------------------------------------------------------------
  Average Duration:           %s
  Min Duration:               %s
  Max Duration:               %s
  P50 (Median):               %s
  P95 (95th Percentile):      %s
  P99 (99th Percentile):      %s
  Total Test Duration:        %s

================================================================================

`,
		m.OperationType,
		m.TotalOperations, m.SuccessCount, m.SuccessRate, m.FailureCount,
		float64(m.TotalBytes)/1_000_000,
		m.OverallThroughputMBps, m.AvgThroughput, m.MinThroughput, m.MaxThroughput,
		formatDuration(m.AvgDuration), formatDuration(m.MinDuration), formatDuration(m.MaxDuration),
		formatDuration(m.P50Duration), formatDuration(m.P95Duration), formatDuration(m.P99Duration),
		formatDuration(m.TotalDuration))
}

// FormatAsCompactTable returns a compact single-line table format
func (m *AggregatedMetrics) FormatAsCompactTable() string {
	return fmt.Sprintf(`
%-12s │ Ops: %5d │ Success: %6.2f%% │ Total: %9.2f MB │ Overall: %7.2f MB/s │ Avg: %10s │ P99: %10s
`, m.OperationType, m.TotalOperations, m.SuccessRate,
		float64(m.TotalBytes)/1_000_000, m.OverallThroughputMBps,
		formatDuration(m.AvgDuration), formatDuration(m.P99Duration))
}

// FormatFileMetric returns a formatted single file metric string
func FormatFileMetric(fileName string, success bool, duration time.Duration, bytes int64) string {
	status := "✓"
	if !success {
		status = "✗"
	}

	mbps := calculateThroughput(bytes, duration)
	mb := float64(bytes) / 1_000_000

	return fmt.Sprintf("  [%s] %s | %.3f MB | %s | %.2f MB/s", status, fileName, mb, formatDuration(duration), mbps)
}
