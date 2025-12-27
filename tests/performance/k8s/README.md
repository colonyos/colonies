# Kubernetes Performance Testing

Run performance benchmarks using Kubernetes for realistic horizontal scaling tests.

## Prerequisites

- Kubernetes cluster (minikube, kind, or real cluster)
- kubectl configured
- Helm 3 installed
- Colonies Helm chart

## Quick Start

```bash
# Run with 1 replica, 1 CPU
./run_k8s_benchmark.sh 1 1000m 1000 10

# Run with 3 replicas, 1 CPU each
./run_k8s_benchmark.sh 3 1000m 1000 10
```

## Scaling Comparison

Compare throughput across different replica counts:

```bash
./run_scaling_comparison.sh
```

This will test with 1, 2, and 3 replicas and generate a comparison summary.

Customize with environment variables:

```bash
PROCESSES=500 EXECUTORS=20 CPU=2000m ./run_scaling_comparison.sh
```

Defaults: 1000 processes, 50 executors, 1000m CPU per replica.

Results saved to `results/scaling_<timestamp>/`.

## Configuration

Edit `values-perf-test.yaml` to customize:

```yaml
# Number of server replicas
ColoniesServerReplicas: 1

# CPU limit per replica (e.g., "500m", "1000m", "2000m")
ColoniesServerCPU: "1000m"

# Memory limit per replica
ColoniesServerMemory: "2000Mi"

# Enable resource limits (must be true for CPU limits to apply)
ColoniesResourceLimit: true
```

## Test Scenarios

### 1. Baseline (1 replica, 1 CPU)

```bash
./run_k8s_benchmark.sh 1 1000m 1000 50
```

### 2. Horizontal Scaling (N replicas, 1 CPU each)

```bash
./run_k8s_benchmark.sh 1 1000m 1000 50
./run_k8s_benchmark.sh 2 1000m 1000 50
./run_k8s_benchmark.sh 3 1000m 1000 50
```

If assign scales horizontally, you should see throughput increase with replicas.

### 3. Vertical Scaling (1 replica, N CPUs)

```bash
./run_k8s_benchmark.sh 1 1000m 1000 50
./run_k8s_benchmark.sh 1 2000m 1000 50
./run_k8s_benchmark.sh 1 4000m 1000 50
```

### 4. High Concurrency

```bash
./run_k8s_benchmark.sh 3 2000m 5000 100
```

## Understanding Results

### Key Metrics

- **Throughput**: Successful assigns per second
- **Latency**: Time from assign request to response
  - P50: Median latency
  - P95: 95th percentile (tail latency)
  - P99: 99th percentile (worst case)

### Expected Behavior

**If leader-only assignment (current):**
- Adding replicas should NOT significantly improve throughput
- All assign requests are forwarded to leader

**If distributed assignment (after optimization):**
- Throughput should scale roughly linearly with replicas
- Each replica handles its own assigns via DB transactions

## Cleanup

```bash
helm uninstall colonies-perf -n colonies-perf
kubectl delete namespace colonies-perf
```

## Troubleshooting

### Check pod status
```bash
kubectl get pods -n colonies-perf
kubectl describe pod <pod-name> -n colonies-perf
```

### View logs
```bash
kubectl logs -n colonies-perf -l app=colonies --tail=100
```

### Check resource usage
```bash
kubectl top pods -n colonies-perf
```
