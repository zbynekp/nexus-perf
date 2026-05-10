#!/bin/bash
# test-baseline.sh - Run a baseline performance test against a Nexus repository.
#
# Configuration is read from NEXUS_PERF_* environment variables (the same
# prefix the binary uses), with sane defaults for optional parameters.
#
# Usage:
#   export NEXUS_PERF_NEXUS_ENDPOINT=https://nexus.example.com
#   export NEXUS_PERF_USERNAME=admin
#   export NEXUS_PERF_PASSWORD=secret
#   export NEXUS_PERF_REPOSITORY_NAME=test-repo
#   ./test-baseline.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/nexus-perf"

if [ ! -f "$BINARY" ]; then
    echo "Building nexus-perf..."
    (cd "$SCRIPT_DIR" && make build)
fi

echo "=================================================="
echo "Running Baseline Performance Test"
echo "=================================================="
echo ""

# Read from NEXUS_PERF_* vars (same prefix as the binary) with defaults.
NEXUS_ENDPOINT="${NEXUS_PERF_NEXUS_ENDPOINT:-https://nexus.example.com}"
NEXUS_USERNAME="${NEXUS_PERF_USERNAME:-admin}"
NEXUS_PASSWORD="${NEXUS_PERF_PASSWORD:-admin123}"
NEXUS_REPO="${NEXUS_PERF_REPOSITORY_NAME:-test-repo}"
NEXUS_FORMAT="${NEXUS_PERF_FORMAT:-RAW}"
NUM_FILES="${NEXUS_PERF_NUM_FILES:-100}"
NUM_THREADS="${NEXUS_PERF_NUM_THREADS:-8}"
FILE_SIZE="${NEXUS_PERF_FILE_SIZE:-1000000}"  # 1 MB (SI: 1 000 000 bytes)
VERBOSITY="${NEXUS_PERF_VERBOSITY:-1}"
LOG_FILE="${NEXUS_PERF_LOG_FILE:-}"
VERBOSE_METRICS="${NEXUS_PERF_VERBOSE_METRICS:-false}"

echo "Test Configuration:"
echo "  Nexus Endpoint: $NEXUS_ENDPOINT"
echo "  Repository:     $NEXUS_REPO ($NEXUS_FORMAT)"
echo "  Files:          $NUM_FILES x $FILE_SIZE bytes"
echo "  Threads:        $NUM_THREADS"
echo "  Verbosity:      $VERBOSITY"
echo ""

ARGS=(
    --nexus-endpoint "$NEXUS_ENDPOINT"
    --username       "$NEXUS_USERNAME"
    --password       "$NEXUS_PASSWORD"
    --repository-name "$NEXUS_REPO"
    --format         "$NEXUS_FORMAT"
    --num-files      "$NUM_FILES"
    --num-threads    "$NUM_THREADS"
    --file-size      "$FILE_SIZE"
    --verbosity      "$VERBOSITY"
)

if [ -n "$LOG_FILE" ]; then
    ARGS+=(--log-file "$LOG_FILE")
fi

if [ "$VERBOSE_METRICS" = "true" ]; then
    ARGS+=(--verbose-metrics)
fi

"$BINARY" test "${ARGS[@]}"

echo ""
echo "Test completed successfully!"
