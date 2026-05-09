# Nexus Performance Test Tool

Enterprise-grade performance testing tool for Sonatype Nexus Repository with support for multi-threaded upload/download tests, comprehensive metrics collection, and flexible configuration.

## Features

- **Multi-threaded Testing**: Concurrent upload and download operations
- **Performance Metrics**: Throughput (Mbps), duration, percentile latencies (P50, P95, P99)
- **Repository Format Support**: RAW and MAVEN repositories
- **TLS/SSL Support**: Custom CA certificates and verification options
- **Flexible Configuration**: CLI parameters and environment variables
- **Automatic Cleanup**: Removes uploaded files after testing (optional)
- **Enterprise Features**:
  - Structured logging with verbosity levels
  - Comprehensive error handling and recovery
  - Thread-safe concurrent operations
  - Graceful degradation on failures
  - Per-file and aggregated metrics

## Installation

### Prerequisites

- Go 1.21 or later

### Build

```bash
cd nexus-perf
go mod download
go build -o nexus-perf
```

### Install

```bash
go install
```

## Configuration

Configuration is managed through three sources (in priority order):

1. **CLI Flags** (highest priority)
2. **Environment Variables** (medium priority)
3. **Defaults** (lowest priority)

### CLI Flags

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password password123 \
  --repository-name test-repo \
  --format RAW \
  --num-files 50 \
  --num-threads 8 \
  --file-size 5242880 \
  --verbosity 2
```

### Environment Variables

All configuration can be set via environment variables with the `NEXUS_PERF_` prefix:

```bash
export NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
export NEXUS_PERF_USERNAME=admin
export NEXUS_PERF_PASSWORD=password123
export NEXUS_PERF_REPOSITORY_NAME=test-repo
export NEXUS_PERF_FORMAT=RAW
export NEXUS_PERF_NUM_FILES=50
export NEXUS_PERF_NUM_THREADS=8
export NEXUS_PERF_FILE_SIZE=5242880
export NEXUS_PERF_VERBOSITY=2

nexus-perf test
```

### Configuration Parameters

| Parameter | Env Variable | Default | Description |
|-----------|--------------|---------|-------------|
| `--nexus-endpoint` | `NEXUS_PERF_NEXUS_ENDPOINT` | *required* | Nexus server URL (e.g., https://nexus.example.com) |
| `--username` | `NEXUS_PERF_USERNAME` | *required* | Nexus username |
| `--password` | `NEXUS_PERF_PASSWORD` | *required* | Nexus password |
| `--repository-name` | `NEXUS_PERF_REPOSITORY_NAME` | *required* | Repository name to test |
| `--format` | `NEXUS_PERF_FORMAT` | `RAW` | Repository format: `RAW` or `MAVEN` |
| `--ca-path` | `NEXUS_PERF_CA_PATH` | (empty) | Path to custom CA certificate file |
| `--skip-verify` | `NEXUS_PERF_SKIP_VERIFY` | `false` | Skip TLS certificate verification |
| `--file-size` | `NEXUS_PERF_FILE_SIZE` | `1048576` | Test file size in bytes (1MB default) |
| `--num-files` | `NEXUS_PERF_NUM_FILES` | `10` | Number of test files to generate |
| `--num-threads` | `NEXUS_PERF_NUM_THREADS` | `4` | Number of concurrent threads |
| `--verbosity` | `NEXUS_PERF_VERBOSITY` | `1` | Verbosity level: 0=error, 1=info, 2=debug, 3=trace |
| `--skip-upload` | `NEXUS_PERF_SKIP_UPLOAD` | `false` | Skip upload test |
| `--skip-download` | `NEXUS_PERF_SKIP_DOWNLOAD` | `false` | Skip download test |
| `--keep-files` | `NEXUS_PERF_KEEP_FILES` | `false` | Keep uploaded files after test |

## Usage

### Full Test (Upload + Download)

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com:8081 \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --format RAW \
  --num-files 100 \
  --num-threads 16 \
  --file-size 10485760 \
  --verbosity 2
```

### Upload Only

```bash
nexus-perf upload \
  --nexus-endpoint https://nexus.example.com:8081 \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --num-files 50 \
  --num-threads 8 \
  --file-size 5242880
```

### Download Only

```bash
nexus-perf download \
  --nexus-endpoint https://nexus.example.com:8081 \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --num-files 50 \
  --num-threads 8
```

### With Custom CA Certificate

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --ca-path /path/to/ca-cert.pem \
  --num-files 50 \
  --num-threads 8
```

### Keep Files for Later Testing

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --num-files 100 \
  --keep-files
```

## Metrics Output

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

=== DOWNLOAD Metrics ===
Total Operations:       100
Success:                100 (100.00%)
Failures:               0
Total Data:             476.84 MB
Total Duration:         38.456s

Throughput:
  Overall:              12.41 Mbps
  Average:              12.18 Mbps
  Min:                  10.23 Mbps
  Max:                  14.67 Mbps

Duration:
  Average:              384ms
  Min:                  318ms
  Max:                  521ms
  P50:                  371ms
  P95:                  489ms
  P99:                  512ms
```

## Repository Format Support

### RAW Repository

- Simple binary file storage
- File paths: `test-files/file-001.bin`, `test-files/file-002.bin`, etc.
- Content-Type: `application/octet-stream`

### MAVEN Repository

- Maven-compatible path structure
- File paths follow Maven convention: `com/example/test/test-artifact-001/1.0.0/test-artifact-001-1.0.0.jar`
- Content-Type: `application/java-archive` for `.jar`, `application/xml` for `.pom`

## Performance Tuning

### Thread Count Optimization

- **Default**: 4 threads
- **Low-latency networks**: Use 16-32 threads
- **High-latency networks**: Use 4-8 threads
- **Guidelines**: Start with 1 thread per 2 CPU cores

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --num-threads 32  # High concurrency
```

### File Size Optimization

- **Small files** (< 1MB): Good for testing latency, use more threads
- **Large files** (> 100MB): Good for testing throughput, use fewer threads
- **Sweet spot**: 5-50MB per file for balanced testing

```bash
# Test with 10MB files
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --file-size 10485760  # 10MB
```

### Network Optimization

- Use dedicated test repository to avoid interference
- Test during low-traffic periods
- Monitor network bandwidth separately
- Consider compression at network level if available

### Connection Pooling

- The tool automatically uses HTTP connection pooling
- Max idle connections: 100
- Max connections per host: 100
- Connection timeout: 30 seconds

## Best Practices

### 1. **Test Planning**
   - Start with small scale (10 files, 2 threads)
   - Gradually increase concurrency
   - Use multiple test runs and average results
   - Test during different times of day

### 2. **Network Considerations**
   - Disable TLS verification only in test environments
   - Use custom CA certificates for production
   - Monitor bandwidth during tests
   - Consider test repository location relative to test client

### 3. **Nexus Configuration**
   - Use dedicated test repository
   - Monitor Nexus server resources during tests
   - Check storage capacity before running
   - Review cleanup logs for completion

### 4. **Security**
   - Never commit passwords to version control
   - Use environment variables for sensitive data
   - Rotate test credentials regularly
   - Audit uploaded files

### 5. **Results Interpretation**
   - Run multiple iterations for stable baselines
   - P95 and P99 latencies indicate worst-case performance
   - Compare throughput at same thread count
   - Account for network variation
   - Track trends over time

### 6. **Troubleshooting**
   - Start with `--verbosity 2` for debugging
   - Check Nexus server logs for errors
   - Verify network connectivity independently
   - Test with smaller file sizes first
   - Review error messages carefully

## Examples

### Baseline Performance Test

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name perf-test \
  --format RAW \
  --num-files 100 \
  --num-threads 8 \
  --file-size 1048576 \
  --verbosity 1
```

### High-concurrency Stress Test

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name stress-test \
  --format RAW \
  --num-files 500 \
  --num-threads 64 \
  --file-size 5242880 \
  --verbosity 2
```

### Large File Throughput Test

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name throughput-test \
  --format RAW \
  --num-files 10 \
  --num-threads 4 \
  --file-size 104857600 \
  --verbosity 1
```

### Maven Repository Test

```bash
nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name maven-releases \
  --format MAVEN \
  --num-files 50 \
  --num-threads 8 \
  --file-size 5242880
```

### Upload Only (Preserve Files)

```bash
nexus-perf upload \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password admin123 \
  --repository-name my-repo \
  --num-files 100 \
  --num-threads 16 \
  --keep-files
```

## Architecture

### Components

1. **config**: Configuration management with validation
2. **logger**: Structured logging with verbosity levels
3. **data**: Random data generation for both RAW and MAVEN formats
4. **nexus**: Nexus repository client with TLS support
5. **metrics**: Performance metrics collection and aggregation
6. **suite**: Test orchestration and execution
7. **cmd**: Command-line interface using Cobra

### Concurrency Model

- Thread-per-worker pattern with configurable pool size
- Work-stealing queue for load balancing
- Buffered channels for synchronization
- Atomic operations for thread-safe metrics

### Error Handling

- Graceful degradation on individual operation failures
- Per-operation error tracking and reporting
- Comprehensive logging at multiple levels
- Cleanup operations even on test failures

## Limitations and Future Enhancements

### Current Limitations

- No persistent result storage
- No scheduling or cron support
- No comparison with baseline results
- Single-node testing only

### Future Enhancements

1. **Result Storage**
   - Export metrics to CSV/JSON
   - Database integration for trend analysis
   - Historical comparison reports

2. **Advanced Testing**
   - Warm-up iterations before metrics collection
   - Latency distribution histograms
   - Network simulation (latency injection)
   - Load curve testing

3. **Monitoring Integration**
   - Prometheus metrics export
   - Graphite integration
   - Custom webhook notifications

4. **Advanced Features**
   - Test templates and profiles
   - Distributed testing across multiple clients
   - Real artifact testing with checksums
   - Bandwidth throttling

## Troubleshooting

### Connection Refused

```
Error: connection refused at https://nexus.example.com:8081
```

**Solutions**:
- Verify Nexus server is running
- Check network connectivity
- Verify correct endpoint URL
- Check firewall rules

### Authentication Failed

```
Error: upload failed with status 401
```

**Solutions**:
- Verify username and password
- Check user has repository permissions
- Verify repository name
- Check if user is locked

### TLS Certificate Error

```
Error: x509: certificate signed by unknown authority
```

**Solutions**:
- Use `--ca-path` with correct CA certificate
- Verify CA certificate is valid PEM format
- Use `--skip-verify` for testing only (NOT recommended for production)

### Out of Memory

```
runtime: out of memory
```

**Solutions**:
- Reduce `--num-files`
- Reduce `--file-size`
- Reduce `--num-threads`
- Increase system memory or use streaming in future version

## License

Proprietary - Contact vendor for licensing information

## Support

For issues, questions, or contributions, please contact the development team.
