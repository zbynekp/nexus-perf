package suite

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/company/nexus-perf/config"
	"github.com/company/nexus-perf/data"
	"github.com/company/nexus-perf/logger"
	"github.com/company/nexus-perf/metrics"
	"github.com/company/nexus-perf/nexus"
)

// TestSuite manages the test execution
type TestSuite struct {
	config *config.Config
	client *nexus.Client
	logger *logger.Logger
	files  []FileInfo
}

// FileInfo stores information about a test file
type FileInfo struct {
	Path string
	Size int64
}

// xferResult is the outcome of a single upload or download
type xferResult struct {
	duration   time.Duration
	bytes      int64
	statusCode int
	err        error
}

// NewTestSuite creates a new test suite
func NewTestSuite(cfg *config.Config, client *nexus.Client, log *logger.Logger) *TestSuite {
	return &TestSuite{
		config: cfg,
		client: client,
		logger: log,
		files:  make([]FileInfo, 0, cfg.NumFiles),
	}
}

// Prepare generates test file paths
func (ts *TestSuite) Prepare() error {
	ts.logger.Infof("Preparing %d test files with size %d bytes each...", ts.config.NumFiles, ts.config.FileSize)

	for i := 0; i < ts.config.NumFiles; i++ {
		ts.files = append(ts.files, FileInfo{
			Path: data.GenerateFilePath(ts.config.Format, i+1),
			Size: ts.config.FileSize,
		})
	}

	ts.logger.Infof("Generated %d test files", len(ts.files))
	return nil
}

func (ts *TestSuite) RunUploadTest(ctx context.Context) (*metrics.AggregatedMetrics, error) {
	return ts.runTransfers(ctx, metrics.OperationUpload, func(ctx context.Context, f FileInfo, t *progressTracker) xferResult {
		r := ts.client.Upload(ctx, f.Path, f.Size, t.wrapReader(io.LimitReader(rand.Reader, f.Size)))
		return xferResult{r.Duration, r.Bytes, r.StatusCode, r.Error}
	})
}

func (ts *TestSuite) RunDownloadTest(ctx context.Context) (*metrics.AggregatedMetrics, error) {
	return ts.runTransfers(ctx, metrics.OperationDownload, func(ctx context.Context, f FileInfo, t *progressTracker) xferResult {
		r := ts.client.Download(ctx, f.Path, t.writer())
		return xferResult{r.Duration, r.Bytes, r.StatusCode, r.Error}
	})
}

// runTransfers is the shared engine for upload and download tests.
// do is called once per file in a goroutine and describes the actual transfer.
func (ts *TestSuite) runTransfers(
	ctx context.Context,
	op metrics.OperationType,
	do func(context.Context, FileInfo, *progressTracker) xferResult,
) (*metrics.AggregatedMetrics, error) {
	ts.logger.Infof("Starting %s test with %d threads, %d files...", op, ts.config.NumThreads, len(ts.files))

	collector := metrics.NewCollector()
	var wg sync.WaitGroup
	sem := make(chan struct{}, ts.config.NumThreads)
	var fileIndex int32
	verboseLines := make([]string, len(ts.files))

	tracker := newProgressTracker(int64(len(ts.files)) * ts.config.FileSize)
	stopProgress := func() {}
	if ts.config.Verbosity <= 1 {
		stopProgress = tracker.run(ctx, string(op))
	}

	for {
		idx := atomic.AddInt32(&fileIndex, 1) - 1
		if idx >= int32(len(ts.files)) || ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			file := ts.files[i]
			ts.logger.Debugf("[%s] file %d/%d: %s", op, i+1, len(ts.files), file.Path)

			r := do(ctx, file, tracker)
			if r.err != nil {
				ts.logger.Warnf("[%s] failed: %s: %v", op, file.Path, r.err)
				collector.Record(r.duration, r.bytes, false, r.statusCode)
				if ts.config.VerboseMetrics {
					verboseLines[i] = metrics.FormatFileMetric(file.Path, false, r.duration, 0)
				}
			} else {
				ts.logger.Debugf("[%s] OK: %s in %v (%.2f MB/s)", op, file.Path, r.duration, calculateMbps(r.bytes, r.duration))
				collector.Record(r.duration, r.bytes, true, r.statusCode)
				if ts.config.VerboseMetrics {
					verboseLines[i] = metrics.FormatFileMetric(file.Path, true, r.duration, r.bytes)
				}
			}
		}(int(idx))
	}

	wg.Wait()
	stopProgress()

	if ts.config.VerboseMetrics {
		for _, line := range verboseLines {
			if line != "" {
				fmt.Println(line)
			}
		}
	}

	if ctx.Err() != nil {
		ts.logger.Infof("%s test interrupted", op)
	} else {
		ts.logger.Infof("%s test completed", op)
	}

	return collector.GetAggregated(op), nil
}

// Cleanup deletes uploaded files. Pass context.Background() to ensure cleanup
// always runs even after the test context is cancelled.
func (ts *TestSuite) Cleanup(ctx context.Context) error {
	if ts.config.KeepFiles {
		ts.logger.Infof("Keeping uploaded files as requested")
		return nil
	}

	ts.logger.Infof("Cleaning up %d files from repository...", len(ts.files))

	var wg sync.WaitGroup
	sem := make(chan struct{}, ts.config.NumThreads)
	var (
		errorCount int
		mu         sync.Mutex
	)

	for i, file := range ts.files {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ts.logger.Debugf("Deleting file %d: %s", idx+1, path)
			if err := ts.client.Delete(ctx, path); err != nil {
				ts.logger.Warnf("Failed to delete %s: %v", path, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
			} else {
				ts.logger.Debugf("Deleted %s", path)
			}
		}(i, file.Path)
	}

	wg.Wait()

	if errorCount > 0 {
		ts.logger.Warnf("Cleanup completed with %d errors", errorCount)
		return fmt.Errorf("cleanup completed with %d errors", errorCount)
	}

	ts.logger.Infof("Cleanup completed successfully")
	return nil
}

func calculateMbps(bytes int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return float64(bytes) / 1_000_000 / duration.Seconds()
}
