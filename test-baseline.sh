#!/bin/bash
# test-baseline.sh - Run baseline performance test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/nexus-perf"

if [ ! -f "$BINARY" ]; then
    echo "Building nexus-perf..."
    cd "$SCRIPT_DIR"
    go build -o nexus-perf
fi

echo "=================================================="
echo "Running Baseline Performance Test"
echo "=================================================="
echo ""

# Configuration
NEXUS_ENDPOINT="${NEXUS_ENDPOINT:-https://nexus.example.com}"
USERNAME="${NEXUS_USERNAME:-admin}"
PASSWORD="${NEXUS_PASSWORD:-admin123}"
REPOSITORY="${NEXUS_REPO:-test-repo}"
FORMAT="${NEXUS_FORMAT:-RAW}"
NUM_FILES="${NUM_FILES:-100}"
NUM_THREADS="${NUM_THREADS:-8}"
FILE_SIZE="${FILE_SIZE:-1048576}"  # 1MB
VERBOSITY="${VERBOSITY:-1}"

echo "Test Configuration:"
echo "  Nexus Endpoint: $NEXUS_ENDPOINT"
echo "  Repository: $REPOSITORY ($FORMAT)"
echo "  Number of Files: $NUM_FILES"
echo "  Number of Threads: $NUM_THREADS"
echo "  File Size: $FILE_SIZE bytes"
echo "  Verbosity: $VERBOSITY"
echo ""

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
echo "Test completed successfully!"
