# Architecture and Design

This document describes the architecture, design patterns, and implementation details of nexus-perf.

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                    nexus-perf CLI                        │
│                   (cmd/root.go)                          │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        v            v            v
    ┌────────┐  ┌──────────┐  ┌──────────┐
    │ Config │  │ Logger   │  │ Nexus    │
    │ Manager│  │ (zap)    │  │ Client   │
    └────────┘  └──────────┘  └──────────┘
        │            │            │
        └────────────┼────────────┘
                     │
                     v
        ┌─────────────────────────┐
        │   Test Suite (Async)    │
        │  - Prepare              │
        │  - RunUploadTest        │
        │  - RunDownloadTest      │
        │  - Cleanup              │
        └────────────┬────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
        v                         v
    ┌──────────────┐        ┌──────────────┐
    │ Data         │        │ Metrics      │
    │ Generator    │        │ Collector    │
    │ (Random)     │        │ (Thread-safe)│
    └──────────────┘        └──────────────┘
```

## Module Structure

### 1. `config` Package
**Responsibility:** Configuration management and validation

**Components:**
- `RepositoryFormat`: Enum for repository types
- `Config`: Main configuration struct
- `New()`: Creates config from environment variables
- `Validate()`: Validates all required parameters
- `ApplyFlags()`: Applies CLI flags to config

**Design Pattern:** Singleton with Builder Pattern
```go
type Config struct {
    NexusEndpoint  string
    Username       string
    Password       string
    // ... other fields
}

func New() (*Config, error) {
    // Load from env, apply defaults, validate
}
```

**Configuration Priority:**
1. CLI Flags (highest)
2. Environment Variables
3. Defaults (lowest)

### 2. `logger` Package
**Responsibility:** Structured logging with verbosity levels

**Components:**
- `Logger`: Wrapper around zap.SugaredLogger
- `New(verbosity)`: Creates logger with appropriate level

**Logging Levels:**
- 0: ERROR - Only errors
- 1: INFO - General information (default)
- 2: DEBUG - Detailed debugging info
- 3: TRACE - Very detailed (same as DEBUG in current impl)

**Design Pattern:** Factory Pattern
```go
func New(verbosity int) *Logger {
    // Create logger based on verbosity level
}
```

### 3. `data` Package
**Responsibility:** Test data generation and file path formatting

**Components:**
- `RandomData()`: Generates cryptographically random bytes
- `GenerateFilePath()`: Creates paths based on repository format
- `generateRawFilePath()`: RAW format paths
- `generateMavenFilePath()`: Maven format paths
- `GetContentType()`: Returns appropriate MIME type

**Format Support:**

**RAW Format:**
```
test-files/file-001.bin
test-files/file-002.bin
...
test-files/file-NNN.bin
```

**Maven Format:**
```
com/example/test/test-artifact-001/1.0.0/test-artifact-001-1.0.0.jar
com/example/test/test-artifact-002/1.0.0/test-artifact-002-1.0.0.jar
```

**Design Pattern:** Strategy Pattern
- Different path generation strategies per format
- Content type determination based on format

### 4. `nexus` Package
**Responsibility:** HTTP client for Nexus API operations

**Components:**
- `Client`: HTTP client with TLS support
- `UploadResult`: Result metrics for upload
- `DownloadResult`: Result metrics for download
- `Upload()`: Uploads file via PUT
- `Download()`: Downloads file via GET
- `Delete()`: Deletes file via DELETE

**HTTP Configuration:**
- Base URL: `https://nexus/repository/{repo}/{path}`
- Authentication: Basic Auth
- Connection pooling: 100 max idle, 100 per host
- Timeout: 30 seconds
- TLS: Custom CA support, skip verify option

**Design Pattern:** Client Pattern
```go
type Client struct {
    endpoint   string
    repository string
    username   string
    password   string
    httpClient *http.Client  // Reusable with pooling
}
```

**API Operations:**
```
Upload:   PUT /repository/{repo}/{path}
Download: GET /repository/{repo}/{path}
Delete:   DELETE /repository/{repo}/{path}
```

### 5. `metrics` Package
**Responsibility:** Performance metrics collection and aggregation

**Components:**
- `ResultMetrics`: Single operation metrics
- `AggregatedMetrics`: Aggregated statistics
- `Collector`: Thread-safe metrics collection
- `Record()`: Records single operation
- `GetAggregated()`: Calculates percentiles and aggregates

**Metrics Collected:**
- Duration (per operation and aggregated)
- Bytes transferred
- Throughput in Mbps
- Success/failure count
- Status codes

**Statistical Calculations:**
- Average, Min, Max
- Percentiles: P50, P95, P99
- Success Rate (%)
- Overall Throughput

**Design Pattern:** Observer/Collector Pattern
```go
type Collector struct {
    mu      sync.RWMutex
    metrics []ResultMetrics  // Thread-safe storage
}
```

### 6. `suite` Package
**Responsibility:** Test orchestration and execution

**Components:**
- `TestSuite`: Main test orchestrator
- `FileInfo`: Metadata for test files
- `Prepare()`: Generates test files
- `RunUploadTest()`: Concurrent upload testing
- `RunDownloadTest()`: Concurrent download testing
- `Cleanup()`: Removes uploaded files

**Concurrency Model:**
```
Worker Pool Pattern:
┌─────────────────────────────────────┐
│        Main Goroutine               │
│    (Spawns N worker goroutines)     │
└────────────┬────────────────────────┘
             │
    ┌────────┴────────┬────────┐
    │                 │        │
    v                 v        v
┌────────┐         ┌────────┐ ┌────────┐
│Worker 1│         │Worker2 │ │Worker N│
└────────┘         └────────┘ └────────┘
    │                 │        │
    └────────┬────────┴────────┘
             │
    ┌────────v─────────────┐
    │  Metrics Collector   │
    │  (Thread-safe via    │
    │   sync.RWMutex)      │
    └─────────────────────┘
```

**Synchronization:**
- Work-stealing queue: `fileIndex int32` with `atomic.AddInt32`
- Semaphore: Buffered channel `sem` limits concurrent operations
- Aggregation: Metrics collected in thread-safe collector

**Design Pattern:** Worker Pool Pattern

### 7. `cmd` Package
**Responsibility:** CLI interface and command routing

**Components:**
- `rootCmd`: Root command
- `testCmd`: Full test (upload + download)
- `uploadCmd`: Upload only
- `downloadCmd`: Download only
- `init()`: Registers flags and bindings
- `runTest()`: Main test execution logic

**CLI Flags:**
- Nexus connection: endpoint, username, password, repository, format
- TLS: ca-path, skip-verify
- Test params: num-files, num-threads, file-size
- Behavior: skip-upload, skip-download, keep-files, verbosity

**Design Pattern:** Command Pattern (via Cobra)

## Concurrency and Thread Safety

### Thread-Safe Operations

1. **Metrics Collection**
```go
// Thread-safe recording
collector.Record(duration, bytes, success, statusCode)
// Implementation uses sync.RWMutex
```

2. **File Index Distribution**
```go
// Atomic counter for work distribution
idx := atomic.AddInt32(&fileIndex, 1) - 1
```

3. **Semaphore for Concurrency Control**
```go
// Buffered channel as semaphore
sem := make(chan struct{}, numThreads)
sem <- struct{}{}        // Acquire
<-sem                    // Release
```

### Race Condition Prevention

- All shared state protected by mutexes or atomics
- Go's memory model ensures happens-before relationships
- Channels used for synchronization
- No global mutable state

## Error Handling Strategy

### Levels of Error Handling

1. **Individual Operation Failures**
   - Per-file upload/download failures recorded
   - Error logged but test continues
   - Success rate metrics reflect failures

2. **Configuration Errors**
   - Caught at initialization
   - Clear error messages for user
   - Application exits with error code

3. **Connection Errors**
   - Retried at HTTP level (no explicit retry)
   - Failure recorded in metrics
   - Test continues with other files

4. **Cleanup Errors**
   - Non-fatal, warns but continues
   - Summary of cleanup failures provided

### Error Recovery

- Graceful degradation on failures
- Partial test results reported
- Cleanup attempts even if tests fail
- No data corruption or incomplete operations

## Performance Considerations

### Memory Efficiency

```go
// Pre-allocated slices
durations := make([]time.Duration, 0, len(c.metrics))

// Streamed operations (potential future)
// Currently loads full file into memory
data := data.GenerateRandomFile(cfg.FileSize)
```

### CPU Efficiency

- Concurrent operations via goroutines
- Minimal lock contention (only for metrics)
- Efficient channel operations
- No busy-waiting

### Network Efficiency

```go
// HTTP connection pooling
Transport: &http.Transport{
    MaxIdleConns:       100,
    MaxIdleConnsPerHost: 10,
    MaxConnsPerHost:    100,
}

// Connection reuse across requests
// Timeout: 30 seconds per operation
```

## Data Flow Diagrams

### Upload Flow
```
┌──────────────┐
│ Generate     │
│ Random Data  │
└──────┬───────┘
       │
       v
┌──────────────────────┐
│ Prepare File Info    │
│ - Path (format-spec) │
│ - Data (random)      │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Worker Threads       │
│ (num-threads pool)   │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ HTTP PUT Request     │
│ /repository/{repo}/{} │
│ Basic Auth           │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Record Metrics       │
│ - Duration           │
│ - Bytes              │
│ - Status Code        │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Nexus Repository     │
│ (File Stored)        │
└──────────────────────┘
```

### Download Flow
```
┌──────────────────────┐
│ Get File List        │
│ from Prepare Phase   │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Worker Threads       │
│ (num-threads pool)   │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ HTTP GET Request     │
│ /repository/{repo}/{} │
│ Basic Auth           │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Nexus Repository     │
│ (File Retrieved)     │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Record Metrics       │
│ - Duration           │
│ - Bytes              │
│ - Status Code        │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Discard Data         │
│ (In-memory only)     │
└──────────────────────┘
```

### Cleanup Flow
```
┌──────────────────────┐
│ Get File List        │
│ from Prepare Phase   │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Worker Threads       │
│ (num-threads pool)   │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ HTTP DELETE Request  │
│ /repository/{repo}/{} │
│ Basic Auth           │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Nexus Repository     │
│ (File Deleted)       │
└──────┬───────────────┘
       │
       v
┌──────────────────────┐
│ Log Cleanup Result   │
│ - Success/Failure    │
│ - Count              │
└──────────────────────┘
```

## Design Patterns Used

### 1. **Singleton Pattern**
- Configuration (single instance)
- Logger (single instance)

### 2. **Factory Pattern**
- Logger creation with verbosity
- Client creation with TLS options

### 3. **Strategy Pattern**
- Data path generation (RAW vs Maven)
- Content-type determination

### 4. **Observer/Collector Pattern**
- Metrics collection and aggregation

### 5. **Command Pattern**
- CLI command routing (Cobra)

### 6. **Worker Pool Pattern**
- Concurrent test execution
- Goroutine pool management

### 7. **Builder Pattern**
- Configuration assembly from multiple sources

## Scalability Considerations

### Current Limitations
- Single-process execution
- Files loaded in memory
- No persistent result storage
- Maximum practical threads: ~1000

### Future Scalability Enhancements
1. **Distributed Testing**
   - Multi-machine test coordination
   - Result aggregation

2. **Streaming**
   - Large file handling with streams
   - Reduced memory footprint

3. **Persistence**
   - Database for historical results
   - Trend analysis capabilities

4. **Advanced Concurrency**
   - Context cancellation
   - Graceful shutdown
   - Rate limiting

## Security Considerations

### Implemented
- TLS certificate validation
- Basic authentication
- Custom CA certificate support
- No hardcoded credentials

### Best Practices
- Use environment variables for secrets
- Rotate test credentials regularly
- Use dedicated test user account
- Audit test operations

### Not Implemented
- Credential rotation
- Audit logging to external system
- API key authentication
- OAuth2 support

## Testing Strategy

### Recommended Test Scenarios
1. **Baseline**: Small files, low concurrency
2. **Stress**: High concurrency, medium files
3. **Throughput**: Large files, low concurrency
4. **Endurance**: Extended duration tests

### Metrics to Track
- Upload/Download throughput (Mbps)
- Latency percentiles (P50, P95, P99)
- Success rate
- Connection stability

## Future Improvements

### High Priority
1. Warm-up iterations before metrics
2. Result export (CSV, JSON)
3. Baseline comparison
4. Performance regression detection

### Medium Priority
1. Network simulation (latency injection)
2. Real artifact testing with checksums
3. Bandwidth throttling
4. Prometheus metrics export

### Low Priority
1. Web UI for test management
2. Scheduled testing
3. Distributed testing
4. Advanced reporting
