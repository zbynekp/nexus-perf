#!/bin/bash
# test-high-concurrency.sh - Run high-concurrency stress test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/nexus-perf"

if [ ! -f "$BINARY" ]; then
    echo "Building nexus-perf..."
    cd "$SCRIPT_DIR"
    go build -o nexus-perf
fi

echo "=================================================="
echo "Running High-Concurrency Stress Test"
echo "=================================================="
echo ""

# Configuration
NEXUS_ENDPOINT="${NEXUS_ENDPOINT:-https://nexus.example.com}"
USERNAME="${NEXUS_USERNAME:-admin}"
PASSWORD="${NEXUS_PASSWORD:-admin123}"
REPOSITORY="${NEXUS_REPO:-stress-test}"
FORMAT="${NEXUS_FORMAT:-RAW}"
NUM_FILES="${NUM_FILES:-500}"
NUM_THREADS="${NUM_THREADS:-64}"
FILE_SIZE="${FILE_SIZE:-5242880}"  # 5MB
VERBOSITY="${VERBOSITY:-2}"

echo "Test Configuration:"
echo "  Nexus Endpoint: $NEXUS_ENDPOINT"
echo "  Repository: $REPOSITORY ($FORMAT)"
echo "  Number of Files: $NUM_FILES"
echo "  Number of Threads: $NUM_THREADS (HIGH CONCURRENCY)"
echo "  File Size: $((FILE_SIZE / 1048576)) MB"
echo "  Verbosity: $VERBOSITY"
echo ""

echo "WARNING: This is a high-concurrency test. Ensure Nexus server has adequate resources."
echo "Press Ctrl+C to cancel, Enter to continue..."
read -r

"$BINARY" test \
    --nexus-endpoint "$NEXUS_ENDPOINT" \
    --username "$USERNAME" \
    --password "$PASSWORD" \
    --repository-name "$REPOSITORY" \
    --format "$FORMAT" \
    --num-files "$NUM_FILES" \
    --num-threads "$NUM_THREADS" \
    --file-size "$FILE_SIZE" \
    --verbosity "$VERBOSITY"

echo ""
echo "Stress test completed!"
