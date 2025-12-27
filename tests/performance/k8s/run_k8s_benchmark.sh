#!/bin/bash
#
# Run performance benchmark on Kubernetes
#
# Prerequisites:
#   - kubectl configured
#   - Helm 3 installed
#   - colonies Helm chart available
#
# Usage:
#   ./run_k8s_benchmark.sh [replicas] [cpu_per_replica] [processes] [executors]
#
# Examples:
#   ./run_k8s_benchmark.sh 1 1000m 1000 10    # 1 replica, 1 CPU
#   ./run_k8s_benchmark.sh 3 1000m 1000 10    # 3 replicas, 1 CPU each
#

set -e

REPLICAS="${1:-1}"
CPU="${2:-1000m}"
PROCESSES="${3:-1000}"
EXECUTORS="${4:-10}"

NAMESPACE="colonies-perf"
RELEASE_NAME="colonies-perf"
HELM_CHART_PATH="${HELM_CHART_PATH:-/home/johan/dev/github/colonyspace/deployments/colonyos.io/helm/colonies}"
VALUES_FILE="$(dirname $0)/values-perf-test.yaml"
KEYS_FILE="$(dirname $0)/.test-keys"
RESULTS_DIR="$(dirname $0)/results"
OUTPUT_DIR="$RESULTS_DIR/$(date +%Y%m%d_%H%M%S)_r${REPLICAS}_cpu${CPU}"

COLONY_NAME="perf-test"

# Generate or load test keys
if [ -f "$KEYS_FILE" ]; then
    echo "Loading existing test keys from $KEYS_FILE"
    source "$KEYS_FILE"
else
    echo "Generating new test keys..."

    # Generate server key
    KEY_OUTPUT=$(colonies security generate 2>&1)
    SERVER_ID=$(echo "$KEY_OUTPUT" | grep -oP 'Id=\K[a-f0-9]+')
    SERVER_PRVKEY=$(echo "$KEY_OUTPUT" | grep -oP 'PrvKey=\K[a-f0-9]+')

    # Generate colony key
    KEY_OUTPUT=$(colonies security generate 2>&1)
    COLONY_ID=$(echo "$KEY_OUTPUT" | grep -oP 'Id=\K[a-f0-9]+')
    COLONY_PRVKEY=$(echo "$KEY_OUTPUT" | grep -oP 'PrvKey=\K[a-f0-9]+')

    if [ -z "$SERVER_ID" ] || [ -z "$SERVER_PRVKEY" ] || [ -z "$COLONY_PRVKEY" ]; then
        echo "Failed to generate keys. Output was:"
        echo "$KEY_OUTPUT"
        exit 1
    fi

    # Save keys for reuse
    cat > "$KEYS_FILE" << EOF
# Auto-generated test keys - do not commit
SERVER_ID="$SERVER_ID"
SERVER_PRVKEY="$SERVER_PRVKEY"
COLONY_ID="$COLONY_ID"
COLONY_PRVKEY="$COLONY_PRVKEY"
EOF
    echo "Saved test keys to $KEYS_FILE"
fi

echo "Server ID: $SERVER_ID"
echo "Colony ID: $COLONY_ID"

echo "=== Colonies K8s Performance Benchmark ==="
echo "Replicas: $REPLICAS"
echo "CPU per replica: $CPU"
echo "Processes: $PROCESSES"
echo "Executors: $EXECUTORS"
echo "Output: $OUTPUT_DIR"
echo ""

mkdir -p "$RESULTS_DIR"
mkdir -p "$OUTPUT_DIR"

# Clean up any existing deployment and keys
echo "Cleaning up any existing deployment..."
helm uninstall $RELEASE_NAME -n $NAMESPACE 2>/dev/null || true
kubectl delete namespace $NAMESPACE 2>/dev/null || true
rm -f "$KEYS_FILE"

# Wait for namespace to be fully deleted
echo "Waiting for cleanup..."
kubectl wait --for=delete namespace/$NAMESPACE --timeout=60s 2>/dev/null || true

# Create namespace
kubectl create namespace $NAMESPACE

# Deploy
echo "Deploying Colonies with $REPLICAS replicas..."
helm upgrade --install $RELEASE_NAME "$HELM_CHART_PATH" \
    --namespace $NAMESPACE \
    -f "$VALUES_FILE" \
    --set ColoniesServerReplicas=$REPLICAS \
    --set ColoniesServerCPU=$CPU \
    --set ColoniesServerID=$SERVER_ID \
    --set ColoniesServerPrvKey=$SERVER_PRVKEY \
    --wait \
    --timeout 5m

# Wait for pods to be ready
echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod \
    -l app=colonies \
    -n $NAMESPACE \
    --timeout=300s

# Get node port
NODE_PORT=$(kubectl get svc colonies-service -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

echo "Colonies available at: $NODE_IP:$NODE_PORT"

# Wait a bit for everything to stabilize
echo "Waiting 10s for cluster to stabilize..."
sleep 10

# Create colony and executor setup (first time only)
echo "Setting up colony..."

# Build benchmark if needed
BENCHMARK_DIR="$(dirname $0)/.."
if [ ! -f "$BENCHMARK_DIR/assign_benchmark" ]; then
    echo "Building benchmark..."
    (cd "$BENCHMARK_DIR" && go build -o assign_benchmark .)
fi

# Run benchmark
echo ""
echo "Running benchmark..."
"$BENCHMARK_DIR/assign_benchmark" \
    --host "$NODE_IP" \
    --port "$NODE_PORT" \
    --insecure \
    --colony "$COLONY_NAME" \
    --colonyprvkey "$COLONY_PRVKEY" \
    --serverprvkey "$SERVER_PRVKEY" \
    --setup \
    --executors "$EXECUTORS" \
    --processes "$PROCESSES" \
    --output "$OUTPUT_DIR/results.csv" \
    --cleanup=true

# Collect pod metrics
echo ""
echo "Collecting pod metrics..."
kubectl top pods -n $NAMESPACE > "$OUTPUT_DIR/pod_metrics.txt" 2>/dev/null || echo "metrics-server not available"

# Save test configuration
cat > "$OUTPUT_DIR/config.txt" << EOF
Test Configuration
==================
Date: $(date)
Replicas: $REPLICAS
CPU per replica: $CPU
Processes: $PROCESSES
Executors: $EXECUTORS
Node: $NODE_IP:$NODE_PORT
EOF

echo ""
echo "=== Results saved to $OUTPUT_DIR ==="
ls -la "$OUTPUT_DIR"
