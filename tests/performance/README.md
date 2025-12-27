# Colonies Assign Performance Test

This benchmark measures the throughput and latency of the assign operation.

## Prerequisites

Start the database and server before running benchmarks:

```bash
# Terminal 1: Start database
make startdb

# Terminal 2: Start server
source devenv
colonies server start
```

## Building

```bash
cd tests/performance
go build -o assign_benchmark .
```

## Running

### Basic Usage

```bash
# Using environment variables (recommended)
source devenv  # or set COLONIES_* variables
./assign_benchmark --executors 10 --processes 1000

# With explicit flags
./assign_benchmark \
  --host localhost \
  --port 50080 \
  --colony dev \
  --colonyprvkey "your-colony-private-key" \
  --executors 10 \
  --processes 1000 \
  --output results.csv
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| --host | localhost | Colonies server host |
| --port | 50080 | Colonies server port |
| --insecure | false | Use HTTP instead of HTTPS |
| --colony | dev | Colony name |
| --colonyprvkey | (required) | Colony private key |
| --executors | 10 | Number of concurrent executors |
| --processes | 1000 | Number of processes to submit |
| --queue-depth | 100 | Wait for this many waiting processes before starting |
| --executor-type | perf-test | Executor type for the test |
| --output | benchmark_results.csv | Output CSV file |
| --cleanup | true | Cleanup after test |

## Limiting CPU for Horizontal Scaling Tests

To simulate the effect of horizontal scaling, limit the server's CPU:

### Option 1: taskset (pin to specific cores)

```bash
# Single core
taskset -c 0 colonies server start

# Two cores
taskset -c 0,1 colonies server start
```

### Option 2: GOMAXPROCS (limit Go threads)

```bash
GOMAXPROCS=1 colonies server start
```

### Option 3: Docker (limit container CPU)

```bash
# Single CPU
docker run --cpus=1 colonyos/colonies server start

# Half a CPU
docker run --cpus=0.5 colonyos/colonies server start
```

### Option 4: cgroups v2

```bash
# Create cgroup with 1 CPU limit
sudo mkdir /sys/fs/cgroup/colonies-test
echo "100000 100000" | sudo tee /sys/fs/cgroup/colonies-test/cpu.max
echo $$ | sudo tee /sys/fs/cgroup/colonies-test/cgroup.procs
colonies server start
```

## Test Scenarios

### 1. Single Node Baseline

```bash
# Terminal 1: Start server with 1 CPU
taskset -c 0 colonies server start

# Terminal 2: Run benchmark
./assign_benchmark --executors 1 --processes 1000 --output single_1exec.csv
./assign_benchmark --executors 10 --processes 1000 --output single_10exec.csv
./assign_benchmark --executors 50 --processes 1000 --output single_50exec.csv
```

### 2. Multi-Node Comparison

```bash
# Terminal 1: Start first server on CPU 0
taskset -c 0 colonies server start --port 50080

# Terminal 2: Start second server on CPU 1
taskset -c 1 colonies server start --port 50081

# Terminal 3: Run benchmark against load balancer or alternate between servers
./assign_benchmark --executors 50 --processes 1000 --port 50080 --output node1.csv
./assign_benchmark --executors 50 --processes 1000 --port 50081 --output node2.csv
```

### 3. Varying Queue Depth

```bash
./assign_benchmark --executors 10 --processes 1000 --queue-depth 10 --output depth_10.csv
./assign_benchmark --executors 10 --processes 1000 --queue-depth 100 --output depth_100.csv
./assign_benchmark --executors 10 --processes 1000 --queue-depth 1000 --output depth_1000.csv
```

## Output Files

The benchmark produces two files:

1. **benchmark_results.csv** - Detailed per-assign results
   - timestamp, executor_id, latency_ms, success

2. **benchmark_results_summary.csv** - Aggregated statistics
   - executors, processes, success/fail counts, latency percentiles

## Analyzing Results

```python
import pandas as pd
import matplotlib.pyplot as plt

# Load results
df = pd.read_csv('benchmark_results.csv')

# Plot latency distribution
df[df['success'] == True]['latency_ms'].hist(bins=50)
plt.xlabel('Latency (ms)')
plt.ylabel('Count')
plt.title('Assign Latency Distribution')
plt.savefig('latency_dist.png')

# Calculate throughput over time
df['timestamp'] = pd.to_datetime(df['timestamp'])
df.set_index('timestamp', inplace=True)
throughput = df[df['success'] == True].resample('1S').count()['latency_ms']
throughput.plot()
plt.xlabel('Time')
plt.ylabel('Assigns/sec')
plt.title('Throughput Over Time')
plt.savefig('throughput.png')
```
