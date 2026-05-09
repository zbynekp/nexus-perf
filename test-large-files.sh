#!/bin/bash
# test-large-files.sh - Run large file throughput test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/nexus-perf"

if [ ! -f "$BINARY" ]; then
    echo "Building nexus-perf..."
    cd "$SCRIPT_DIR"
    go build -o nexus-perf
fi

echo "=================================================="
echo "Running Large File Throughput Test"
echo "=================================================="
echo ""

# Configuration
NEXUS_ENDPOINT="${NEXUS_ENDPOINT:-https://nexus.example.com}"
USERNAME="${NEXUS_USERNAME:-admin}"
PASSWORD="${NEXUS_PASSWORD:-admin123}"
REPOSITORY="${NEXUS_REPO:-throughput-test}"
FORMAT="${NEXUS_FORMAT:-RAW}"
NUM_FILES="${NUM_FILES:-10}"
NUM_THREADS="${NUM_THREADS:-4}"
FILE_SIZE="${FILE_SIZE:-104857600}"  # 100MB
VERBOSITY="${VERBOSITY:-1}"

echo "Test Configuration:"
echo "  Nexus Endpoint: $NEXUS_ENDPOINT"
echo "  Repository: $REPOSITORY ($FORMAT)"
echo "  Number of Files: $NUM_FILES"
echo "  Number of Threads: $NUM_THREADS (LOW CONCURRENCY FOR THROUGHPUT)"
echo "  File Size: $((FILE_SIZE / 1048576)) MB"
echo "  Total Data: $((NUM_FILES * FILE_SIZE / 1048576)) MB"
echo "  Verbosity: $VERBOSITY"
echo ""

echo "WARNING: This test will transfer $((NUM_FILES * FILE_SIZE / 1048576)) MB total data."
echo "Ensure adequate network bandwidth is available."
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
echo "Throughput test completed!"
