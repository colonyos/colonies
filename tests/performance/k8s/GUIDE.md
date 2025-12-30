# Colonies K8s Performance Benchmark Guide

## Running Benchmarks

### Basic Usage
```bash
./run_k8s_benchmark.sh [replicas] [cpu_per_replica] [processes] [executors]

# Examples:
./run_k8s_benchmark.sh 1 1000m 1000 10    # 1 replica, 1 CPU
./run_k8s_benchmark.sh 3 1000m 50000 10   # 3 replicas, 1 CPU each
```

### Environment Variables
```bash
METRICS_INTERVAL=1       # Metrics collection interval in seconds (default: 1)
USE_PROMETHEUS=true      # Collect app metrics from Prometheus (default: true)
PROMETHEUS_PORT=9090     # Local port for Prometheus port-forward (default: 9090)
```

### Scaling Comparison
```bash
./run_scaling_comparison.sh [processes] [executors]

# Runs benchmarks with 1, 3, 5, 7, 9, 11 replicas
./run_scaling_comparison.sh 50000 10
```

## Output Files

Each benchmark run creates a timestamped directory with:

| File | Description |
|------|-------------|
| `results.csv` | Per-process timing data |
| `results_summary.csv` | Aggregated metrics (throughput, latency) |
| `pod_metrics_timeseries.csv` | CPU/memory per pod over time |
| `app_metrics.csv` | Colonies process states over time |
| `colonies_logs.jsonl` | Server logs with timestamps |
| `config.txt` | Test configuration |

## Plotting Results

### Basic Plots
```bash
# Plot all experiments
python3 plot_results.py

# Plot latest experiment only
python3 plot_results.py --latest

# List available experiments
python3 plot_results.py --list

# Publication-quality output (PDF + PNG)
python3 plot_results.py --publication
```

### Specific Plot Types
```bash
# CPU timeseries for specific replica count
python3 plot_results.py --timeseries 3 --latest

# App metrics (processes waiting/running/completed)
python3 plot_results.py --app-metrics 3 --latest

# CPU with log event density overlay
python3 plot_results.py --cpu-logs 3 --latest
```

## Syncing CPU with Colonies Logs

### Automatic Log Collection
The benchmark script automatically collects timestamped logs:
```bash
./run_k8s_benchmark.sh 3 1000m 50000 10
# Creates: results/.../colonies_logs.jsonl
```

### Correlation Analysis
```bash
# Analyze CPU spikes and show surrounding log context
python3 correlate_logs_cpu.py results/scaling_*/replicas_3/

# Options:
#   --window 2      Seconds of log context around spikes (default: 2)
#   --threshold 90  CPU percentile for spike detection (default: 90)
#   --output FILE   Output file for combined timeline
```

Output includes:
- CPU spikes with log entries within the time window
- Log pattern analysis (Assign, Close, Submit, errors)
- Combined timeline CSV (`cpu_logs_timeline.csv`)

### Visualization
```bash
# Plot CPU with log event density overlay
python3 plot_results.py --cpu-logs 3 --latest
```

This shows:
- CPU utilization (primary y-axis)
- Log events per second (secondary y-axis, bar chart)

### Timestamp Alignment

Both data sources use Unix timestamps for easy correlation:

| Source | File | Timestamp Format |
|--------|------|------------------|
| CPU metrics | `pod_metrics_timeseries.csv` | Unix seconds (column: `timestamp`) |
| Logs | `colonies_logs.jsonl` | ISO 8601 from `kubectl logs --timestamps` |

## High-Resolution Metrics

### Prometheus Collector
For sub-second metrics when cadvisor is available:
```bash
python3 collect_prometheus_metrics.py \
    --prometheus http://localhost:9090 \
    --namespace colonies-perf \
    --output metrics.csv \
    --app-metrics app_metrics.csv \
    --interval 1 \
    --duration 120
```

### Metrics Sources

| Metric Type | Source | Resolution |
|-------------|--------|------------|
| Container CPU/Memory | kubectl top (metrics-server) | ~1 second |
| Container CPU/Memory | Prometheus + cadvisor | configurable |
| App metrics (process counts) | Prometheus + Colonies monitor | configurable |

## Configuration

### values-perf-test.yaml

Key settings for performance testing:

```yaml
# Server scaling
ColoniesServerReplicas: 1
ColoniesResourceLimit: false     # Set true to enable CPU throttling
ColoniesServerCPU: "500m"
ColoniesServerCPULimit: "500m"

# Database
DBResourceLimit: false
DBCPU: "8000m"
ColoniesDBMaxOpenConns: 100
ColoniesDBMaxIdleConns: 100

# Monitoring
ScrapeIntervall: 5s              # Prometheus scrape interval
```

### CPU Throttling Behavior

When `ColoniesResourceLimit: true`:
- Server is limited to specified CPU
- Under high load, operations queue up
- Latency increases significantly
- Useful for testing scaling behavior

When `ColoniesResourceLimit: false`:
- Server can use all available CPU
- Lower latency under load
- Useful for finding peak performance

## Troubleshooting

### Benchmark Hangs
- Check for stuck processes: `colonies process ps`
- Verify database connectivity
- Check server logs for errors

### Missing Metrics
- Ensure metrics-server is running: `kubectl top nodes`
- Check Prometheus is scraping: access Prometheus UI

### High Latency
- Check CPU throttling (resource limits)
- Check database CPU usage
- Increase replicas or CPU limits
