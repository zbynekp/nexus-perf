# Quick Start Guide

Get up and running with nexus-perf in 5 minutes!

## Prerequisites

- Go 1.21+ (for building from source)
- Access to Nexus Repository Manager 3.x
- Basic command-line knowledge

## 1. Build the Application

```bash
# Clone or download the project
cd nexus-perf

# Build
go build -o nexus-perf

# Or use Make
make build
```

## 2. Prepare Nexus

### Create a Test Repository

1. Log in to Nexus web console (http://localhost:8081)
2. Click "Administration" → "Repositories"
3. Click "Create Repository"
4. Select "Raw Repository"
5. Name: `nexus-perf-test`
6. Click "Create Repository"

### Create a Test User

1. Go to "Administration" → "Security" → "Users"
2. Click "Create User"
3. Username: `testuser`
4. Password: `testpass123`
5. Email: `test@example.com`
6. Assign role with repository privileges
7. Click "Create User"

## 3. Run Your First Test

```bash
# Set up environment variables
export NEXUS_PERF_NEXUS_ENDPOINT=http://localhost:8081
export NEXUS_PERF_USERNAME=testuser
export NEXUS_PERF_PASSWORD=testpass123
export NEXUS_PERF_REPOSITORY_NAME=nexus-perf-test

# Run a simple test
./nexus-perf test \
  --num-files 10 \
  --num-threads 2 \
  --file-size 1048576 \
  --verbosity 1
```

## 4. Understand the Output

```
=== UPLOAD Metrics ===
Total Operations:       10
Success:                10 (100.00%)
Failures:               0
Total Data:             9.54 MB
Total Duration:         5.234s

Throughput:
  Overall:              1.82 Mbps
  Average:              1.78 Mbps
  Min:                  1.52 Mbps
  Max:                  2.15 Mbps

Duration:
  Average:              523ms
  Min:                  385ms
  Max:                  687ms
  P50:                  515ms
  P95:                  651ms
  P99:                  678ms
```

**What this means:**
- Successfully uploaded 10 files (1MB each)
- Average speed: ~1.8 Mbps
- P50 (median) upload time: 515ms
- P95 (95th percentile) upload time: 651ms

## 5. Try Different Scenarios

### Test Upload Only

```bash
./nexus-perf upload \
  --num-files 20 \
  --num-threads 4 \
  --file-size 5242880
```

### Test Download Only

```bash
./nexus-perf download \
  --num-files 20 \
  --num-threads 4
```

### Test with Different File Sizes

```bash
# Large files (measure throughput)
./nexus-perf test \
  --num-files 5 \
  --num-threads 2 \
  --file-size 104857600  # 100MB
```

### Test with High Concurrency

```bash
# Many threads (stress test)
./nexus-perf test \
  --num-files 100 \
  --num-threads 32 \
  --file-size 5242880
```

## 6. Using Configuration Files

### Create `.env` File

```bash
cat > .env << EOF
NEXUS_PERF_NEXUS_ENDPOINT=http://localhost:8081
NEXUS_PERF_USERNAME=testuser
NEXUS_PERF_PASSWORD=testpass123
NEXUS_PERF_REPOSITORY_NAME=nexus-perf-test
NEXUS_PERF_FORMAT=RAW
NEXUS_PERF_NUM_FILES=50
NEXUS_PERF_NUM_THREADS=8
NEXUS_PERF_VERBOSITY=1
EOF

# Load and run
source .env
./nexus-perf test
```

## 7. Docker Quick Start

```bash
# Build Docker image
docker build -t nexus-perf:latest .

# Run with Docker Compose (includes local Nexus)
docker-compose up

# Or run standalone
docker run --rm -it \
  -e NEXUS_PERF_NEXUS_ENDPOINT=http://nexus.example.com \
  -e NEXUS_PERF_USERNAME=admin \
  -e NEXUS_PERF_PASSWORD=admin123 \
  -e NEXUS_PERF_REPOSITORY_NAME=test-repo \
  nexus-perf:latest test
```

## 8. Common Tasks

### Save Results to File

```bash
./nexus-perf test --num-files 50 --num-threads 8 | tee results-$(date +%Y%m%d-%H%M%S).log
```

### Keep Uploaded Files

```bash
./nexus-perf test --num-files 50 --keep-files
```

### Use HTTPS with Self-Signed Cert

```bash
# Export CA cert from Nexus
openssl s_client -connect nexus.example.com:8443 | \
  sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > ca.pem

# Use with tool
./nexus-perf test --ca-path ca.pem --nexus-endpoint https://nexus.example.com
```

### Debug Connection Issues

```bash
# Verbose output
./nexus-perf test --num-files 5 --verbosity 3

# Test connectivity
curl -u testuser:testpass123 http://localhost:8081/service/rest/v1/status
```

## 9. Next Steps

- Read [README.md](README.md) for full documentation
- Check [ENTERPRISE_GUIDE.md](ENTERPRISE_GUIDE.md) for enterprise deployment
- Review test scripts in `test-*.sh` for more examples
- Explore performance tuning recommendations

## Troubleshooting

### Connection Refused
```bash
# Check Nexus is running
curl http://localhost:8081
```

### Authentication Failed
```bash
# Verify credentials
curl -u testuser:testpass123 http://localhost:8081/service/rest/v1/status
```

### Permission Denied
```bash
# Ensure user has write permissions on repository
# In Nexus: Administration → Security → Users → testuser → add privileges
```

### Out of Memory
```bash
# Reduce load parameters
./nexus-perf test \
  --num-files 10 \
  --num-threads 2 \
  --file-size 1048576
```

## Need Help?

- Check logs with `--verbosity 3`
- Review error messages carefully
- Consult [README.md](README.md) troubleshooting section
- Enable debug logging for deeper investigation

---

**Ready for advanced testing?** Check out the [ENTERPRISE_GUIDE.md](ENTERPRISE_GUIDE.md)!
