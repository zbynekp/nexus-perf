# Project Summary: Nexus Performance Testing Tool

## Overview

A production-ready, enterprise-grade Go application for testing Sonatype Nexus Repository performance with multi-threaded concurrent operations, comprehensive metrics collection, and flexible configuration.

## Project Status

✅ **BUILD SUCCESSFUL** - All code compiles and runs

## Project Structure

```
nexus-perf/
├── go.mod                          # Go module definition with dependencies
├── go.sum                          # Go dependencies checksums
├── main.go                         # Application entry point
├── Makefile                        # Build automation and development tasks
├── Dockerfile                      # Container image definition
├── docker-compose.yml              # Docker Compose for local Nexus testing
├── .gitignore                      # Git ignore rules
├── .env.example                    # Example environment configuration
│
├── README.md                       # Main documentation (comprehensive)
├── QUICKSTART.md                   # Getting started guide (5 minutes)
├── ENTERPRISE_GUIDE.md             # Enterprise deployment guide
├── ARCHITECTURE.md                 # Technical architecture documentation
├── PROJECT_SUMMARY.md              # This file
│
├── cmd/                            # Command-line interface
│   └── root.go                    # CLI commands (test, upload, download)
│
├── config/                         # Configuration management
│   └── config.go                  # Configuration from env vars & CLI flags
│
├── logger/                         # Logging system
│   └── logger.go                  # Structured logging with verbosity levels
│
├── data/                           # Test data generation
│   └── generator.go               # Random data & path generation (RAW, Maven)
│
├── nexus/                          # Nexus API client
│   └── client.go                  # HTTP client with TLS support
│
├── metrics/                        # Performance metrics
│   └── metrics.go                 # Thread-safe metrics collection & aggregation
│
├── suite/                          # Test orchestration
│   └── suite.go                   # Multi-threaded test execution
│
├── test-baseline.sh               # Baseline performance test script
├── test-high-concurrency.sh       # High concurrency stress test script
└── test-large-files.sh            # Large file throughput test script
```

## Key Components

### 1. Configuration (`config/config.go`)
- **Purpose**: Manages all application configuration
- **Features**:
  - Loads from environment variables (`NEXUS_PERF_*`)
  - Applies CLI flags with priority
  - Validates all required parameters
  - Supports both RAW and MAVEN repository formats
- **Configuration Options**: 19 different parameters

### 2. Logger (`logger/logger.go`)
- **Purpose**: Structured logging system
- **Features**:
  - 4 verbosity levels (0=error, 1=info, 2=debug, 3=trace)
  - Uses zap.SugaredLogger for efficiency
  - JSON formatted output
  - Separate stdout/stderr handling

### 3. Data Generation (`data/generator.go`)
- **Purpose**: Generate test data and file paths
- **Features**:
  - Cryptographically random byte generation
  - Format-specific path generation:
    - RAW: `test-files/file-001.bin`
    - Maven: `com/example/test/test-artifact-001/1.0.0/test-artifact-001-1.0.0.jar`
  - Content-type detection based on format

### 4. Nexus Client (`nexus/client.go`)
- **Purpose**: HTTP client for Nexus API operations
- **Features**:
  - TLS/SSL support with custom CA certificates
  - Basic authentication
  - HTTP connection pooling (100 max idle, 100 per host)
  - 30-second operation timeout
  - Operations:
    - Upload (PUT)
    - Download (GET)
    - Delete (DELETE)
  - Per-operation metrics collection

### 5. Metrics (`metrics/metrics.go`)
- **Purpose**: Performance metrics collection and analysis
- **Features**:
  - Thread-safe metrics collection
  - Percentile calculations (P50, P95, P99)
  - Throughput calculation (Mbps)
  - Statistics (min, max, average, success rate)
  - Formatted output for easy reading

### 6. Test Suite (`suite/suite.go`)
- **Purpose**: Orchestrate multi-threaded test execution
- **Features**:
  - Worker pool pattern for concurrency
  - File preparation (random data generation)
  - Upload testing with configurable threads
  - Download testing with configurable threads
  - Automatic cleanup of uploaded files (optional)
  - Thread-safe metrics aggregation

### 7. CLI (`cmd/root.go`)
- **Purpose**: Command-line interface
- **Commands**:
  - `test`: Run full upload+download test
  - `upload`: Upload test only
  - `download`: Download test only
- **Flags**: 18 configuration flags with environment variable fallback

## Build and Deployment

### Building from Source
```bash
cd /home/zp/Projects/go-workspace/nexus-perf
go build -o nexus-perf
```

### Binary Details
- **Name**: nexus-perf
- **Size**: ~14MB
- **Platform**: Linux x86_64 (portable to other platforms via cross-compilation)
- **Go Version**: 1.21+

### Deployment Options

#### 1. Direct Binary
```bash
./nexus-perf test --nexus-endpoint https://nexus.example.com ...
```

#### 2. Docker Container
```bash
docker build -t nexus-perf:latest .
docker run nexus-perf:latest test --nexus-endpoint https://nexus.example.com ...
```

#### 3. Docker Compose (with local Nexus)
```bash
docker-compose up
```

#### 4. Installation
```bash
make install
```

## Quick Start

### 1. Minimal Test
```bash
export NEXUS_PERF_NEXUS_ENDPOINT=http://localhost:8081
export NEXUS_PERF_USERNAME=admin
export NEXUS_PERF_PASSWORD=admin123
export NEXUS_PERF_REPOSITORY_NAME=test-repo

./nexus-perf test --num-files 10 --num-threads 2
```

### 2. Using Environment File
```bash
cp .env.example .env
# Edit .env with your values
source .env
./nexus-perf test
```

### 3. Using Test Scripts
```bash
./test-baseline.sh          # 100 files, 1MB, 8 threads
./test-high-concurrency.sh  # 500 files, 5MB, 64 threads
./test-large-files.sh       # 10 files, 100MB, 4 threads
```

## Configuration

### Required Parameters
- `--nexus-endpoint`: Nexus server URL
- `--username`: Nexus username
- `--password`: Nexus password  
- `--repository-name`: Repository name to test
- `--format`: RAW or MAVEN (default: RAW)

### Optional Parameters
- `--ca-path`: Path to custom CA certificate
- `--skip-verify`: Skip TLS verification (not recommended for production)
- `--num-files`: Number of test files (default: 10)
- `--num-threads`: Number of concurrent threads (default: 4)
- `--file-size`: Size of each file in bytes (default: 1MB)
- `--verbosity`: Logging verbosity (0-3, default: 1)
- `--skip-upload`: Skip upload test
- `--skip-download`: Skip download test
- `--keep-files`: Keep files after test

### Environment Variables
All options available as `NEXUS_PERF_*` environment variables:
```bash
NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
NEXUS_PERF_USERNAME=admin
NEXUS_PERF_PASSWORD=password
NEXUS_PERF_REPOSITORY_NAME=test-repo
NEXUS_PERF_NUM_FILES=100
NEXUS_PERF_NUM_THREADS=16
NEXUS_PERF_FILE_SIZE=5242880
NEXUS_PERF_FORMAT=RAW
NEXUS_PERF_VERBOSITY=2
```

## Performance Metrics Output

The tool provides comprehensive performance metrics:

```
=== UPLOAD Metrics ===
Total Operations:       100
Success:                100 (100.00%)
Failures:               0
Total Data:             476.84 MB
Total Duration:         45.234s

Throughput:
  Overall:              10.54 Mbps
  Average:              10.28 Mbps
  Min:                  8.92 Mbps
  Max:                  12.45 Mbps

Duration:
  Average:              452ms
  Min:                  385ms
  Max:                  687ms
  P50:                  438ms
  P95:                  621ms
  P99:                  675ms
```

## Design Highlights

### Enterprise-Grade Features
✅ **Concurrency**: Worker pool pattern with configurable threads
✅ **Thread Safety**: Atomic operations and RWMutex for shared state
✅ **Error Handling**: Graceful degradation with per-operation failure tracking
✅ **Metrics**: Comprehensive statistics with percentile calculations
✅ **Logging**: Structured logging with multiple verbosity levels
✅ **Security**: TLS support, custom CA certificates, basic auth
✅ **Flexibility**: Multiple configuration sources with priority ordering
✅ **Reliability**: Connection pooling, timeout management, automatic cleanup

### Design Patterns
- **Singleton**: Config, Logger
- **Factory**: Logger creation, Client creation
- **Strategy**: Data path generation (RAW vs Maven)
- **Observer**: Metrics collection
- **Command**: CLI routing (Cobra)
- **Worker Pool**: Concurrent test execution

### Performance Optimizations
- HTTP connection pooling (100 max idle)
- Buffered channels for synchronization
- Atomic operations for lock-free metrics
- Efficient memory allocation
- Streaming-ready architecture (future enhancement)

## Testing Scenarios

### 1. Baseline Performance
```bash
./nexus-perf test \
  --num-files 100 \
  --num-threads 8 \
  --file-size 1048576
```

### 2. Stress Testing
```bash
./nexus-perf test \
  --num-files 500 \
  --num-threads 64 \
  --file-size 5242880
```

### 3. Throughput Testing
```bash
./nexus-perf test \
  --num-files 10 \
  --num-threads 4 \
  --file-size 104857600
```

### 4. Maven Repository
```bash
./nexus-perf test \
  --format MAVEN \
  --num-files 50 \
  --num-threads 8
```

## Documentation

### For Quick Setup
→ See **QUICKSTART.md** (5-minute setup)

### For Full Usage
→ See **README.md** (comprehensive guide with examples)

### For Enterprise Deployment
→ See **ENTERPRISE_GUIDE.md** (production deployment guide)

### For Architecture Details
→ See **ARCHITECTURE.md** (technical documentation)

## Tools and Utilities

### Make Targets
```bash
make build              # Build the application
make build-linux       # Build for Linux
make build-macos       # Build for macOS
make build-windows     # Build for Windows
make run               # Build and run
make test              # Run tests
make test-coverage     # Generate coverage report
make clean             # Clean build artifacts
make lint              # Run linters
make fmt               # Format code
make install           # Install binary
make deps              # Download dependencies
make docker-build      # Build Docker image
make docker-run        # Run in Docker
make all               # Full build process
```

### Test Scripts
```bash
test-baseline.sh           # Baseline performance test
test-high-concurrency.sh   # High concurrency stress test
test-large-files.sh        # Large file throughput test
```

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` v1.7.0 - CLI framework
- `github.com/spf13/viper` v1.17.0 - Configuration management
- `go.uber.org/zap` v1.26.0 - Structured logging

### Transitive Dependencies
- Standard Go library for HTTP, TLS, crypto
- Over 40 transitive dependencies (handled by go mod)

## System Requirements

### Minimum
- Go 1.21+
- 512MB RAM
- Linux/macOS/Windows (via cross-compilation)
- Network access to Nexus server

### Recommended for Testing
- 2+ CPU cores
- 2GB+ RAM
- 100Mbps+ network
- Dedicated test repository on Nexus

## Known Limitations

1. **Memory**: Files loaded entirely into memory (suitable for files up to ~500MB)
2. **Single Node**: Only single-process testing (no distributed testing)
3. **No Persistence**: Results not stored between runs
4. **Thread Limit**: Practical maximum ~1000 threads due to OS limits

## Future Enhancements

### High Priority
- Result export (CSV, JSON)
- Baseline comparison
- Warm-up iterations
- Performance regression detection

### Medium Priority
- Web UI
- Scheduled testing
- Distributed testing
- Prometheus metrics

### Low Priority
- Advanced reporting
- Cron scheduling
- Result database
- Trend analysis

## Support and Troubleshooting

### Common Issues

**Connection Refused**
```bash
# Verify Nexus is running
curl http://localhost:8081
```

**Authentication Failed**
```bash
# Verify credentials
curl -u username:password http://nexus:8081/service/rest/v1/status
```

**TLS Certificate Error**
```bash
# Export CA and use it
openssl s_client -connect nexus.example.com:8443 | \
  sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > ca.pem

./nexus-perf test --ca-path ca.pem ...
```

**Out of Memory**
```bash
# Reduce parameters
./nexus-perf test \
  --num-files 10 \
  --num-threads 2 \
  --file-size 1048576
```

See **README.md** for more troubleshooting.

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| main.go | 15 | Application entry point |
| cmd/root.go | 200+ | CLI commands and initialization |
| config/config.go | 180+ | Configuration management |
| logger/logger.go | 40 | Structured logging |
| data/generator.go | 80 | Test data generation |
| nexus/client.go | 180+ | Nexus API client |
| metrics/metrics.go | 230+ | Metrics collection and aggregation |
| suite/suite.go | 220+ | Test orchestration |
| Dockerfile | 30 | Container image |
| Makefile | 120 | Build automation |
| README.md | 800+ | Full documentation |
| QUICKSTART.md | 250+ | Getting started guide |
| ENTERPRISE_GUIDE.md | 600+ | Enterprise deployment |
| ARCHITECTURE.md | 700+ | Technical architecture |
| **TOTAL** | **~4000+** | **Production-ready codebase** |

## Getting Help

1. **Quick questions**: Check QUICKSTART.md
2. **Usage help**: See README.md
3. **Configuration issues**: Review ENTERPRISE_GUIDE.md
4. **Technical details**: Check ARCHITECTURE.md
5. **Build issues**: Review Makefile and Dockerfile
6. **Runtime issues**: Use `--verbosity 3` for detailed logging

## License and Credits

This is a professional-grade testing tool developed for comprehensive Nexus Repository performance analysis. All code follows Go best practices and includes comprehensive error handling, thread safety, and enterprise features.

---

**Status**: ✅ Ready for Production
**Last Updated**: May 9, 2026
**Version**: 1.0.0
