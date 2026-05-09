package suite

import (
	"fmt"
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
	files  []FileInfo // list of files to upload/download
	mu     sync.Mutex
}

// FileInfo stores information about uploaded files
type FileInfo struct {
	Path string
	Data []byte
	Size int64
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

// Prepare generates test files
func (ts *TestSuite) Prepare() error {
	ts.logger.Infof("Preparing %d test files with size %d bytes each...", ts.config.NumFiles, ts.config.FileSize)

	for i := 0; i < ts.config.NumFiles; i++ {
		fileData := data.GenerateRandomFile(ts.config.FileSize)
		filePath := data.GenerateFilePath(ts.config.Format, i+1)

		ts.files = append(ts.files, FileInfo{
			Path: filePath,
			Data: fileData,
			Size: ts.config.FileSize,
		})
	}

	ts.logger.Infof("Generated %d test files", len(ts.files))
	return nil
}

// RunUploadTest runs the upload test with multiple threads
func (ts *TestSuite) RunUploadTest() (*metrics.AggregatedMetrics, error) {
	ts.logger.Infof("Starting upload test with %d threads, %d files...", ts.config.NumThreads, len(ts.files))

	collector := metrics.NewCollector()
	var wg sync.WaitGroup
	var activeWorkers int32
	sem := make(chan struct{}, ts.config.NumThreads)
	fileIndex := int32(0)

	// Start worker goroutines
	for w := 0; w < ts.config.NumThreads; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			atomic.AddInt32(&activeWorkers, 1)
			defer atomic.AddInt32(&activeWorkers, -1)

			for {
				idx := atomic.AddInt32(&fileIndex, 1) - 1
				if idx >= int32(len(ts.files)) {
					break
				}

				sem <- struct{}{}
				go func(fileIdx int) {
					defer func() { <-sem }()

					file := ts.files[fileIdx]
					ts.logger.Debugf("[Worker %d] Uploading file %d: %s (%d bytes)", workerID, fileIdx+1, file.Path, file.Size)

					result := ts.client.Upload(file.Path, file.Data)

					if result.Error != nil {
						ts.logger.Warnf("[Worker %d] Upload failed for %s: %v", workerID, file.Path, result.Error)
						collector.Record(result.Duration, result.Size, false, result.StatusCode)
					} else {
						ts.logger.Debugf("[Worker %d] Upload successful for %s in %v (%.2f Mbps)", workerID, file.Path, result.Duration, calculateMbps(result.Size, result.Duration))
						collector.Record(result.Duration, result.Size, true, result.StatusCode)
					}
				}(int(idx))
			}
		}(w)
	}

	wg.Wait()
	ts.logger.Infof("Upload test completed")

	return collector.GetAggregated(metrics.OperationUpload), nil
}

// RunDownloadTest runs the download test with multiple threads
func (ts *TestSuite) RunDownloadTest() (*metrics.AggregatedMetrics, error) {
	ts.logger.Infof("Starting download test with %d threads, %d files...", ts.config.NumThreads, len(ts.files))

	collector := metrics.NewCollector()
	var wg sync.WaitGroup
	var activeWorkers int32
	sem := make(chan struct{}, ts.config.NumThreads)
	fileIndex := int32(0)

	// Start worker goroutines
	for w := 0; w < ts.config.NumThreads; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			atomic.AddInt32(&activeWorkers, 1)
			defer atomic.AddInt32(&activeWorkers, -1)

			for {
				idx := atomic.AddInt32(&fileIndex, 1) - 1
				if idx >= int32(len(ts.files)) {
					break
				}

				sem <- struct{}{}
				go func(fileIdx int) {
					defer func() { <-sem }()

					file := ts.files[fileIdx]
					ts.logger.Debugf("[Worker %d] Downloading file %d: %s", workerID, fileIdx+1, file.Path)

					result := ts.client.Download(file.Path)

					if result.Error != nil {
						ts.logger.Warnf("[Worker %d] Download failed for %s: %v", workerID, file.Path, result.Error)
						collector.Record(result.Duration, 0, false, result.StatusCode)
					} else {
						ts.logger.Debugf("[Worker %d] Download successful for %s in %v (%.2f Mbps)", workerID, file.Path, result.Duration, calculateMbps(result.Size, result.Duration))
						collector.Record(result.Duration, result.Size, true, result.StatusCode)
					}
				}(int(idx))
			}
		}(w)
	}

	wg.Wait()
	ts.logger.Infof("Download test completed")

	return collector.GetAggregated(metrics.OperationDownload), nil
}

// Cleanup deletes uploaded files if not keeping them
func (ts *TestSuite) Cleanup() error {
	if ts.config.KeepFiles {
		ts.logger.Infof("Keeping uploaded files as requested")
		return nil
	}

	ts.logger.Infof("Cleaning up %d files from repository...", len(ts.files))

	var wg sync.WaitGroup
	sem := make(chan struct{}, ts.config.NumThreads)
	errorCount := 0
	var mu sync.Mutex

	for i, file := range ts.files {
		wg.Add(1)
		go func(fileIdx int, filePath string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			ts.logger.Debugf("Deleting file %d: %s", fileIdx+1, filePath)

			if err := ts.client.Delete(filePath); err != nil {
				ts.logger.Warnf("Failed to delete file %s: %v", filePath, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
			} else {
				ts.logger.Debugf("Deleted file %s", filePath)
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

// calculateMbps calculates throughput in Mbps
func calculateMbps(bytes int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	megabytes := float64(bytes) / (1024 * 1024)
	seconds := duration.Seconds()
	return megabytes / seconds
}
