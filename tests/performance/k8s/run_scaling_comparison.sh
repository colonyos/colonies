#!/bin/bash
#
# Run horizontal scaling comparison
#
# Tests different replica counts with fixed CPU per replica
# to measure how assign throughput scales
#

set -e

PROCESSES="${PROCESSES:-1000}"
EXECUTORS="${EXECUTORS:-50}"
CPU="${CPU:-1000m}"
RESULTS_DIR="$(dirname $0)/results"

# Resume mode: reuse scaling directory with existing results
if [ "${RESUME:-0}" = "1" ]; then
    # Find latest directory that actually has results
    FOUND_DIR=""
    for dir in $(ls -td "$RESULTS_DIR"/scaling_* 2>/dev/null); do
        if ls "$dir"/replicas_*/results_summary.csv >/dev/null 2>&1; then
            FOUND_DIR="$dir"
            break
        fi
    done
    if [ -n "$FOUND_DIR" ]; then
        OUTPUT_DIR="$FOUND_DIR"
        echo "Resuming previous run: $OUTPUT_DIR"
        # Show what's already done
        echo "Already completed:"
        for f in "$OUTPUT_DIR"/replicas_*/results_summary.csv; do
            if [ -f "$f" ]; then
                rep=$(basename $(dirname "$f"))
                echo "  - $rep"
            fi
        done
    else
        OUTPUT_DIR="$RESULTS_DIR/scaling_$(date +%Y%m%d_%H%M%S)"
        echo "No previous results found, starting fresh"
    fi
else
    OUTPUT_DIR="$RESULTS_DIR/scaling_$(date +%Y%m%d_%H%M%S)"
fi

mkdir -p "$RESULTS_DIR"
mkdir -p "$OUTPUT_DIR"

echo "=== Horizontal Scaling Comparison ==="
echo "Processes: $PROCESSES"
echo "Executors: $EXECUTORS"
echo "CPU per replica: $CPU"
echo "Output: $OUTPUT_DIR"
echo ""

# Test with different replica counts
REPLICA_COUNTS="${REPLICA_COUNTS:-1 3 5 7 9}"
for REPLICAS in $REPLICA_COUNTS; do
    # Skip if already done
    if [ -d "$OUTPUT_DIR/replicas_$REPLICAS" ] && [ -f "$OUTPUT_DIR/replicas_$REPLICAS/results_summary.csv" ]; then
        echo ""
        echo "=========================================="
        echo "Skipping $REPLICAS replica(s) - already done"
        echo "=========================================="
        continue
    fi

    echo ""
    echo "=========================================="
    echo "Testing with $REPLICAS replica(s)..."
    echo "=========================================="

    ./run_k8s_benchmark.sh $REPLICAS $CPU $PROCESSES $EXECUTORS

    # Move results
    LATEST_RESULT=$(ls -td "$RESULTS_DIR"/*_r${REPLICAS}_* 2>/dev/null | head -1)
    if [ -n "$LATEST_RESULT" ]; then
        mv "$LATEST_RESULT" "$OUTPUT_DIR/replicas_$REPLICAS"
    fi

    # Brief pause between tests
    echo "Waiting 30s before next test..."
    sleep 30
done

# Generate comparison summary
echo ""
echo "=== Comparison Summary ==="
echo ""

SUMMARY_FILE="$OUTPUT_DIR/comparison_summary.csv"
echo "replicas,successful,failed,avg_latency_ms,p95_latency_ms,p99_latency_ms,throughput_per_sec" > "$SUMMARY_FILE"

for REPLICAS in $REPLICA_COUNTS; do
    RESULT_DIR="$OUTPUT_DIR/replicas_$REPLICAS"
    SUMMARY="$RESULT_DIR/results_summary.csv"

    if [ -f "$SUMMARY" ]; then
        SUCCESS=$(grep "^successful," "$SUMMARY" | cut -d',' -f2)
        FAILED=$(grep "^failed," "$SUMMARY" | cut -d',' -f2)
        AVG=$(grep "^avg_latency_ms," "$SUMMARY" | cut -d',' -f2)
        P95=$(grep "^p95_latency_ms," "$SUMMARY" | cut -d',' -f2)
        P99=$(grep "^p99_latency_ms," "$SUMMARY" | cut -d',' -f2)

        # Calculate approximate throughput (successful / avg_latency * 1000 * executors)
        # This is a rough estimate
        if [ "$AVG" != "0" ] && [ "$AVG" != "" ]; then
            THROUGHPUT=$(echo "scale=2; $SUCCESS / ($AVG / 1000) / $EXECUTORS" | bc 2>/dev/null || echo "N/A")
        else
            THROUGHPUT="N/A"
        fi

        echo "$REPLICAS,$SUCCESS,$FAILED,$AVG,$P95,$P99,$THROUGHPUT" >> "$SUMMARY_FILE"
    fi
done

echo ""
cat "$SUMMARY_FILE" | column -t -s','
echo ""
echo "Results saved to: $OUTPUT_DIR"
