#!/bin/bash
#
# Run comparison benchmark with different executor counts
#
# Usage:
#   ./run_comparison.sh [output_dir]
#
# Prerequisites:
#   - Colonies server running
#   - Environment variables set (source devenv)
#

set -e

OUTPUT_DIR="${1:-results_$(date +%Y%m%d_%H%M%S)}"
mkdir -p "$OUTPUT_DIR"

PROCESSES="${PROCESSES:-1000}"
QUEUE_DEPTH="${QUEUE_DEPTH:-100}"

echo "=== Colonies Assign Performance Comparison ==="
echo "Output directory: $OUTPUT_DIR"
echo "Processes per test: $PROCESSES"
echo "Queue depth: $QUEUE_DEPTH"
echo ""

# Build if needed
if [ ! -f ./assign_benchmark ]; then
    echo "Building benchmark..."
    go build -o assign_benchmark .
fi

# Test with varying executor counts
for EXECUTORS in 1 5 10 25 50; do
    echo ""
    echo "=== Running with $EXECUTORS executors ==="
    ./assign_benchmark \
        --executors $EXECUTORS \
        --processes $PROCESSES \
        --queue-depth $QUEUE_DEPTH \
        --output "$OUTPUT_DIR/exec_${EXECUTORS}.csv" \
        --cleanup=true

    # Brief pause between tests
    sleep 2
done

echo ""
echo "=== Comparison Summary ==="
echo ""
printf "%-10s %-10s %-10s %-10s %-10s %-10s\n" "Executors" "Success" "Failed" "Avg(ms)" "P95(ms)" "P99(ms)"
printf "%-10s %-10s %-10s %-10s %-10s %-10s\n" "---------" "-------" "------" "-------" "-------" "-------"

for EXECUTORS in 1 5 10 25 50; do
    SUMMARY="$OUTPUT_DIR/exec_${EXECUTORS}_summary.csv"
    if [ -f "$SUMMARY" ]; then
        SUCCESS=$(grep "^successful," "$SUMMARY" | cut -d',' -f2)
        FAILED=$(grep "^failed," "$SUMMARY" | cut -d',' -f2)
        AVG=$(grep "^avg_latency_ms," "$SUMMARY" | cut -d',' -f2)
        P95=$(grep "^p95_latency_ms," "$SUMMARY" | cut -d',' -f2)
        P99=$(grep "^p99_latency_ms," "$SUMMARY" | cut -d',' -f2)
        printf "%-10s %-10s %-10s %-10s %-10s %-10s\n" "$EXECUTORS" "$SUCCESS" "$FAILED" "$AVG" "$P95" "$P99"
    fi
done

echo ""
echo "Results saved to $OUTPUT_DIR/"
