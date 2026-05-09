# Enterprise Deployment Guide

This document provides guidelines for deploying and running nexus-perf in enterprise environments.

## Pre-Deployment Checklist

- [ ] Nexus Repository Manager 3.x installed and configured
- [ ] Test repository created in Nexus
- [ ] User account with read/write permissions on test repository
- [ ] Network connectivity verified between test client and Nexus
- [ ] TLS certificates validated (if using HTTPS)
- [ ] Adequate storage space in test repository
- [ ] Nexus server resources monitored during tests
- [ ] Test environment isolated from production workloads
- [ ] Backup of important data completed

## Installation and Setup

### 1. Build from Source

```bash
git clone <repository>
cd nexus-perf
go mod download
make build
```

### 2. Verify Installation

```bash
./nexus-perf --version
./nexus-perf --help
```

### 3. Run Initial Test

```bash
# Test with minimal parameters
export NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
export NEXUS_PERF_USERNAME=testuser
export NEXUS_PERF_PASSWORD=testpass
export NEXUS_PERF_REPOSITORY_NAME=test-repo

./nexus-perf test \
  --num-files 5 \
  --num-threads 1 \
  --file-size 1048576 \
  --verbosity 2
```

## Nexus Configuration for Testing

### Create Test Repository

1. Log in to Nexus web interface
2. Create new repository:
   - Type: Choose based on format (Raw or Maven)
   - Name: e.g., `nexus-perf-test-raw`
   - Public/Private: Private recommended
3. Create dedicated test user:
   - Username: `nexus-perf-test`
   - Password: Use strong, randomly generated password
   - Privileges: Read, Write, Edit on test repository only

### Repository Setup Examples

**Raw Repository:**
```
Name: nexus-perf-test-raw
Type: Raw
Policy: Release
Cleanup Policy: Enable with 1-day retention
```

**Maven Repository:**
```
Name: nexus-perf-test-maven
Type: Maven 2 (Hosted)
Policy: Release
Cleanup Policy: Enable with 1-day retention
```

## Performance Testing Strategy

### Phase 1: Baseline Testing (Week 1)

```bash
# Test 1: Small files, low concurrency
./nexus-perf test \
  --num-files 50 \
  --num-threads 2 \
  --file-size 1048576 \
  --verbosity 1

# Test 2: Small files, medium concurrency
./nexus-perf test \
  --num-files 100 \
  --num-threads 8 \
  --file-size 1048576 \
  --verbosity 1

# Test 3: Medium files, medium concurrency
./nexus-perf test \
  --num-files 50 \
  --num-threads 8 \
  --file-size 5242880 \
  --verbosity 1
```

### Phase 2: Stress Testing (Week 2)

```bash
# Test 4: High concurrency
./nexus-perf test \
  --num-files 200 \
  --num-threads 32 \
  --file-size 5242880 \
  --verbosity 1

# Test 5: Large files
./nexus-perf test \
  --num-files 10 \
  --num-threads 4 \
  --file-size 104857600 \
  --verbosity 1
```

### Phase 3: Load Testing (Week 3)

```bash
# Test 6: Extended duration
./nexus-perf test \
  --num-files 500 \
  --num-threads 16 \
  --file-size 5242880 \
  --keep-files

# Manual verification between upload and download
sleep 300

# Download test on kept files
./nexus-perf download \
  --num-files 500 \
  --num-threads 16
```

## Monitoring During Tests

### System Metrics to Track

1. **Nexus Server:**
   - CPU utilization
   - Memory usage
   - Disk I/O
   - Network throughput
   - Heap memory (JVM)

2. **Test Client:**
   - Network bandwidth usage
   - System memory
   - CPU utilization
   - Connection count

### Monitoring Tools

**Linux:**
```bash
# Monitor in real-time
watch -n 1 'top -b -n 1 | head -20'
watch -n 1 'iostat -x 1 2'
watch -n 1 'netstat -i'
```

**Docker/Kubernetes:**
```bash
# Monitor container
docker stats nexus-perf-test

# Monitor Nexus
docker logs -f nexus-test
```

## Results Collection and Analysis

### Export Metrics

The tool outputs metrics to stdout. Capture for analysis:

```bash
./nexus-perf test \
  --num-files 100 \
  --num-threads 8 \
  --verbosity 1 | tee results-$(date +%Y%m%d-%H%M%S).log
```

### Key Metrics to Track

1. **Throughput (Mbps)**
   - Overall throughput
   - Per-thread throughput
   - Min/Max throughput

2. **Latency (ms)**
   - P50, P95, P99 percentiles
   - Min/Max latency
   - Average latency

3. **Success Rate**
   - Percentage of successful operations
   - Number of failures
   - Error patterns

### Comparative Analysis

Create a tracking spreadsheet:

| Date | Test Type | Files | Threads | File Size | Upload Mbps | Download Mbps | Success Rate |
|------|-----------|-------|---------|-----------|-------------|---------------|--------------|
| 2024-01-01 | Baseline | 50 | 2 | 1MB | 8.5 | 9.2 | 100% |
| 2024-01-02 | Baseline | 100 | 8 | 1MB | 7.9 | 8.8 | 100% |
| 2024-01-05 | Stress | 200 | 32 | 5MB | 6.5 | 7.4 | 99.5% |

## Capacity Planning

### Disk Space Requirements

```
Total Data = Number of Files × File Size × 2 (upload + download)
Example: 100 files × 5MB × 2 = 1GB required
```

### Network Bandwidth

```
Required Bandwidth = (Total Data × 2) / Total Duration
Example: 1GB / 300s ≈ 2.66 Mbps sustained
```

### Memory Requirements

```
Client-side: ~100MB + (Number of Threads × 50MB)
Example: 100MB + (16 × 50MB) = 900MB

Nexus-side: Monitor JVM heap during tests
Typical: 2GB heap minimum for testing
```

## Troubleshooting Guide

### Slow Performance

**Symptoms:**
- Throughput < expected baseline
- High P95/P99 latencies
- Inconsistent results

**Solutions:**
1. Reduce number of threads
2. Reduce file size
3. Check network connectivity
4. Monitor Nexus server resources
5. Verify no other workloads on Nexus

### Connection Failures

**Symptoms:**
```
Error: dial tcp <nexus>:8081: i/o timeout
```

**Solutions:**
1. Verify Nexus is running: `curl -I https://nexus.example.com:8081`
2. Check firewall rules
3. Verify DNS resolution
4. Check network connectivity: `ping nexus.example.com`

### Authentication Failures

**Symptoms:**
```
Error: upload failed with status 401
```

**Solutions:**
1. Verify credentials: `curl -u username:password https://nexus.example.com:8081/service/rest/v1/status`
2. Check user permissions
3. Verify repository name
4. Check if user account is locked

### TLS Certificate Issues

**Symptoms:**
```
Error: x509: certificate signed by unknown authority
```

**Solutions:**
1. Export Nexus CA certificate:
```bash
echo | openssl s_client -connect nexus.example.com:8443 \
  | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > ca.pem
```

2. Use with tool:
```bash
./nexus-perf test --ca-path ca.pem ...
```

### Out of Memory

**Symptoms:**
```
runtime: out of memory: cannot allocate
```

**Solutions:**
1. Reduce `--num-files`
2. Reduce `--file-size`
3. Reduce `--num-threads`
4. Increase system RAM

## Security Best Practices

### Credential Management

```bash
# Use environment files (secure mode: 600)
cat > .env.nexus
NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
NEXUS_PERF_USERNAME=nexus-perf-test
NEXUS_PERF_PASSWORD=<secure-password>
NEXUS_PERF_REPOSITORY_NAME=nexus-perf-test-raw

chmod 600 .env.nexus
source .env.nexus
./nexus-perf test
```

### TLS/SSL

```bash
# Always use HTTPS in production
./nexus-perf test --nexus-endpoint https://nexus.example.com ...

# Validate with custom CA
./nexus-perf test --ca-path /path/to/ca.pem ...

# Never use --skip-verify in production
```

### Access Control

1. Create dedicated test user with limited permissions
2. Use IP allowlisting if possible
3. Rotate credentials regularly
4. Audit test runs
5. Clean up test files after testing

## Cleanup Procedures

### Remove Test Files

```bash
# Files are automatically removed by default
# If kept with --keep-files, manually remove:
./nexus-perf cleanup --num-files 100

# Or manually in Nexus web interface
# Nexus > Repositories > test-repo > Delete files
```

### Reset Test Repository

```bash
# Method 1: Via Nexus web interface
# Administration > Repositories > test-repo > Delete repository
# Then recreate repository

# Method 2: Via API
curl -X DELETE -u admin:password \
  https://nexus.example.com/service/rest/v1/repositories/raw/nexus-perf-test-raw
```

### Archive Results

```bash
# Save results for analysis
mkdir -p results/$(date +%Y-%m-%d)
cp *.log results/$(date +%Y-%m-%d)/
tar czf results-$(date +%Y%m%d).tar.gz results/
```

## Maintenance Schedule

### Daily
- Monitor test repository size
- Check for errors in logs
- Verify Nexus is running

### Weekly
- Review performance metrics
- Archive old results
- Clean up test files

### Monthly
- Update tool to latest version
- Review baseline metrics
- Adjust test parameters if needed

### Quarterly
- Full system capacity assessment
- Performance trend analysis
- Update documentation

## Support and Escalation

### Common Issues Matrix

| Issue | Resolution | Severity |
|-------|-----------|----------|
| Connection timeout | Check network/firewall | HIGH |
| Auth failures | Verify credentials | HIGH |
| Out of memory | Reduce load parameters | HIGH |
| Slow performance | Check server resources | MEDIUM |
| Intermittent errors | Increase P95/P99 thresholds | MEDIUM |
| Certificate errors | Update CA certificates | LOW |

### Contact Information

- Development Team: perf-testing@company.com
- Nexus Support: nexus-support@company.com
- DevOps Team: devops@company.com

## Appendix: Complete Configuration Reference

```bash
# All available configuration options
./nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password password123 \
  --repository-name test-repo \
  --format RAW \                    # or MAVEN
  --ca-path /path/to/ca.pem \
  --skip-verify false \
  --num-files 100 \
  --num-threads 8 \
  --file-size 5242880 \             # 5MB
  --verbosity 2 \
  --skip-upload false \
  --skip-download false \
  --keep-files false
```
