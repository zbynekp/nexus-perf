# nexus-perf

A CLI tool for benchmarking upload and download throughput against a [Sonatype Nexus Repository](https://www.sonatype.com/products/sonatype-nexus-repository) instance. It runs concurrent transfers, measures per-file and aggregate metrics, and prints a summary table with throughput (MB/s), latency percentiles (P50/P95/P99), and success rate.

Supports RAW and Maven repository formats. TLS with a custom CA or skip-verify is available for internal/self-signed setups.

## Requirements

- Go 1.25+ (for building from source)
- A reachable Nexus 3 instance with a repository the user can read/write

## Quick start

```bash
# Build
make build

# Run a basic test
./nexus-perf test \
  --nexus-endpoint https://nexus.example.com \
  --username admin \
  --password secret \
  --repository-name test-repo \
  --num-files 20 \
  --num-threads 4 \
  --file-size 5000000
```

Or via environment variables (useful in CI or Docker):

```bash
export NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
export NEXUS_PERF_USERNAME=admin
export NEXUS_PERF_PASSWORD=secret
export NEXUS_PERF_REPOSITORY_NAME=test-repo

./nexus-perf test
```

Copy `.env.example` to `.env`, fill in your values, and source it:

```bash
cp .env.example .env
# edit .env
set -a; source .env; set +a
./nexus-perf test
```

## Subcommands

| Command | Description |
|---|---|
| `test` | Run upload then download (default) |
| `upload` | Upload only |
| `download` | Download only |

## Key flags

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--nexus-endpoint` | `NEXUS_PERF_NEXUS_ENDPOINT` | *(required)* | Nexus base URL |
| `--username` | `NEXUS_PERF_USERNAME` | *(required)* | Nexus username |
| `--password` | `NEXUS_PERF_PASSWORD` | *(required)* | Nexus password |
| `--repository-name` | `NEXUS_PERF_REPOSITORY_NAME` | *(required)* | Target repository |
| `--format` | `NEXUS_PERF_FORMAT` | `RAW` | `RAW` or `MAVEN` |
| `--num-files` | `NEXUS_PERF_NUM_FILES` | `10` | Number of test files |
| `--num-threads` | `NEXUS_PERF_NUM_THREADS` | `4` | Concurrent transfers |
| `--file-size` | `NEXUS_PERF_FILE_SIZE` | `1000000` | File size in bytes (SI MB) |
| `--verbosity` | `NEXUS_PERF_VERBOSITY` | `1` | 0=error 1=info 2=debug 3=trace |
| `--skip-upload` | `NEXUS_PERF_SKIP_UPLOAD` | `false` | Skip upload phase |
| `--skip-download` | `NEXUS_PERF_SKIP_DOWNLOAD` | `false` | Skip download phase |
| `--keep-files` | `NEXUS_PERF_KEEP_FILES` | `false` | Do not delete files after test |
| `--log-file` | `NEXUS_PERF_LOG_FILE` | *(stdout)* | Write structured JSON logs to file |
| `--verbose-metrics` | `NEXUS_PERF_VERBOSE_METRICS` | `false` | Print per-file timing and throughput |
| `--ca-path` | `NEXUS_PERF_CA_PATH` | — | Custom CA certificate path |
| `--skip-verify` | `NEXUS_PERF_SKIP_VERIFY` | `false` | Skip TLS verification |

CLI flags take precedence over environment variables.

## Docker

```bash
# Build image
make docker-build VERSION=1.0.0

# Run
make docker-run \
  NEXUS_ENDPOINT=https://nexus.example.com \
  NEXUS_USERNAME=admin \
  NEXUS_PASSWORD=secret \
  NEXUS_REPO=test-repo
```

A `docker-compose.yml` is included for spinning up a local Nexus instance alongside the perf tool:

```bash
NEXUS_PERF_PASSWORD=secret docker compose up
```

## Make targets

```
make build          Build for host OS/arch
make build-linux    Cross-compile for Linux amd64
make build-macos    Cross-compile for macOS amd64
make build-windows  Cross-compile for Windows amd64
make test           Run tests
make test-coverage  Run tests with HTML coverage report
make lint           Run go vet + golangci-lint
make fmt            Format code
make clean          Remove build artifacts
make deps           Download and tidy dependencies
make docker-build   Build Docker image
make docker-run     Run in Docker
```

## Baseline script

`test-baseline.sh` is a convenience wrapper that builds the binary if needed and runs the test using `NEXUS_PERF_*` environment variables:

```bash
export NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
export NEXUS_PERF_USERNAME=admin
export NEXUS_PERF_PASSWORD=secret
export NEXUS_PERF_REPOSITORY_NAME=test-repo
./test-baseline.sh
```
