#!/usr/bin/env python3
"""
Plot scaling comparison results from Colonies performance benchmarks.

Usage:
    python3 plot_results.py                    # Plot all experiments
    python3 plot_results.py --latest           # Plot only the latest
    python3 plot_results.py --list             # List available experiments
    python3 plot_results.py <dir1> <dir2> ...  # Plot specific experiments
    python3 plot_results.py --publication      # Publication-quality output
"""

import sys
import os
import glob
import csv
import argparse
import matplotlib.pyplot as plt
import matplotlib as mpl
import numpy as np
from datetime import datetime

def get_script_dir():
    return os.path.dirname(os.path.abspath(__file__))

def find_all_experiments():
    """Find all scaling results directories."""
    results_dir = os.path.join(get_script_dir(), "results")
    scaling_dirs = glob.glob(os.path.join(results_dir, "scaling_*"))
    return sorted(scaling_dirs, key=os.path.getmtime)

def parse_experiment_name(path):
    """Extract a readable name from experiment path."""
    basename = os.path.basename(path)
    if basename.startswith("scaling_"):
        try:
            ts = basename.replace("scaling_", "")
            dt = datetime.strptime(ts, "%Y%m%d_%H%M%S")
            return dt.strftime("%Y-%m-%d %H:%M")
        except:
            pass
    return basename

def read_summary(filepath):
    """Read a results_summary.csv file into a dict."""
    data = {}
    with open(filepath, 'r') as f:
        reader = csv.reader(f)
        for row in reader:
            if len(row) >= 2:
                try:
                    data[row[0]] = float(row[1])
                except ValueError:
                    data[row[0]] = row[1]
    return data

def read_experiment(exp_dir):
    """Read all replica results from an experiment directory."""
    results = {
        'name': parse_experiment_name(exp_dir),
        'path': exp_dir,
        'replicas': [],
        'avg_latency': [],
        'p50_latency': [],
        'p95_latency': [],
        'p99_latency': [],
        'min_latency': [],
        'max_latency': [],
        'avg_cpu': [],
        'max_cpu': [],
        'db_avg_cpu': [],
        'db_max_cpu': [],
        'executors': 0,
        'processes': 0,
    }

    for r in range(1, 15):  # Support up to 14 replicas
        summary_file = os.path.join(exp_dir, f"replicas_{r}", "results_summary.csv")
        if os.path.exists(summary_file):
            data = read_summary(summary_file)
            results['replicas'].append(r)
            results['avg_latency'].append(data.get('avg_latency_ms', 0))
            results['p50_latency'].append(data.get('p50_latency_ms', 0))
            results['p95_latency'].append(data.get('p95_latency_ms', 0))
            results['p99_latency'].append(data.get('p99_latency_ms', 0))
            results['min_latency'].append(data.get('min_latency_ms', 0))
            results['max_latency'].append(data.get('max_latency_ms', 0))
            results['avg_cpu'].append(data.get('avg_cpu_millicores', 0))
            results['max_cpu'].append(data.get('max_cpu_millicores', 0))
            results['db_avg_cpu'].append(data.get('db_avg_cpu_millicores', 0))
            results['db_max_cpu'].append(data.get('db_max_cpu_millicores', 0))
            if results['executors'] == 0:
                results['executors'] = int(data.get('executors', 0))
                results['processes'] = int(data.get('processes', 0))

    return results

def setup_publication_style():
    """Configure matplotlib for publication-quality figures."""
    plt.style.use('seaborn-v0_8-whitegrid')
    mpl.rcParams.update({
        # Font settings - use serif for publications
        'font.family': 'serif',
        'font.serif': ['Times New Roman', 'DejaVu Serif', 'serif'],
        'font.size': 10,
        'axes.labelsize': 11,
        'axes.titlesize': 11,
        'xtick.labelsize': 10,
        'ytick.labelsize': 10,
        'legend.fontsize': 9,
        'figure.titlesize': 12,
        # Line and marker settings
        'lines.linewidth': 1.5,
        'lines.markersize': 6,
        # Grid
        'grid.alpha': 0.3,
        'grid.linestyle': '--',
        # Figure settings
        'figure.dpi': 300,
        'savefig.dpi': 300,
        'savefig.bbox': 'tight',
        'savefig.pad_inches': 0.05,
        # Remove top and right spines
        'axes.spines.top': False,
        'axes.spines.right': False,
    })

# Color palette suitable for publications (colorblind-friendly)
COLORS = {
    'primary': '#0072B2',    # Blue
    'secondary': '#D55E00',  # Orange
    'tertiary': '#009E73',   # Green
    'quaternary': '#CC79A7', # Pink
    'gray': '#666666',
}

def plot_experiment(exp_data, output_file, publication_mode=False):
    """Plot single combined chart for an experiment."""
    replicas = exp_data['replicas']

    # Check if we have resource metrics
    has_metrics = any(exp_data['avg_cpu']) and any(v > 0 for v in exp_data['avg_cpu'])
    has_db_metrics = any(exp_data['db_avg_cpu']) and any(v > 0 for v in exp_data['db_avg_cpu'])

    if publication_mode:
        setup_publication_style()
        # Publication figure size (single column ~3.5in, double column ~7in)
        if has_metrics:
            fig, axes = plt.subplots(2, 2, figsize=(7, 5.5))
        else:
            fig, axes = plt.subplots(1, 2, figsize=(7, 2.75))
    else:
        if has_metrics:
            fig, axes = plt.subplots(2, 2, figsize=(14, 12))
            fig.suptitle(f"Colonies Assign Performance - {exp_data['name']}\n({exp_data['executors']} executors, {exp_data['processes']} processes)",
                         fontsize=14, fontweight='bold')
        else:
            fig, axes = plt.subplots(1, 2, figsize=(14, 6))
            fig.suptitle(f"Colonies Assign Latency - {exp_data['name']}\n({exp_data['executors']} executors, {exp_data['processes']} processes)",
                         fontsize=14, fontweight='bold')

    if has_metrics:
        ax1, ax2, ax3, ax4 = axes[0, 0], axes[0, 1], axes[1, 0], axes[1, 1]
    else:
        ax1, ax2 = axes[0], axes[1]

    # Plot 1: Latency percentiles (bar chart)
    x = np.arange(len(replicas))
    width = 0.2

    if publication_mode:
        bars1 = ax1.bar(x - width*1.5, exp_data['avg_latency'], width, label='Mean', color=COLORS['tertiary'], edgecolor='white', linewidth=0.5)
        bars2 = ax1.bar(x - width*0.5, exp_data['p50_latency'], width, label='P50', color=COLORS['primary'], edgecolor='white', linewidth=0.5)
        bars3 = ax1.bar(x + width*0.5, exp_data['p95_latency'], width, label='P95', color=COLORS['secondary'], edgecolor='white', linewidth=0.5)
        bars4 = ax1.bar(x + width*1.5, exp_data['p99_latency'], width, label='P99', color=COLORS['quaternary'], edgecolor='white', linewidth=0.5)
    else:
        bars1 = ax1.bar(x - width*1.5, exp_data['avg_latency'], width, label='Average', color='#2ecc71')
        bars2 = ax1.bar(x - width*0.5, exp_data['p50_latency'], width, label='P50', color='#3498db')
        bars3 = ax1.bar(x + width*0.5, exp_data['p95_latency'], width, label='P95', color='#e74c3c')
        bars4 = ax1.bar(x + width*1.5, exp_data['p99_latency'], width, label='P99', color='#9b59b6')

    ax1.set_xlabel('Number of Replicas')
    ax1.set_ylabel('Latency (ms)')
    ax1.set_title('(a) Latency Percentiles')
    ax1.set_xticks(x)
    ax1.set_xticklabels([str(r) for r in replicas])
    ax1.legend(loc='upper right', framealpha=0.9)

    if not publication_mode:
        for bars in [bars1, bars2, bars3, bars4]:
            for bar in bars:
                height = bar.get_height()
                ax1.annotate(f'{height:.0f}',
                            xy=(bar.get_x() + bar.get_width() / 2, height),
                            xytext=(0, 3),
                            textcoords="offset points",
                            ha='center', va='bottom', fontsize=8)

    # Plot 2: Latency trend (line chart)
    if publication_mode:
        ax2.fill_between(replicas, exp_data['min_latency'], exp_data['max_latency'],
                         alpha=0.15, color=COLORS['primary'], label='Min-Max')
        ax2.plot(replicas, exp_data['avg_latency'], 'o-', color=COLORS['tertiary'],
                 linewidth=1.5, markersize=5, label='Mean')
        ax2.plot(replicas, exp_data['p95_latency'], 's--', color=COLORS['secondary'],
                 linewidth=1.5, markersize=4, label='P95')
        ax2.plot(replicas, exp_data['p99_latency'], '^:', color=COLORS['quaternary'],
                 linewidth=1.5, markersize=4, label='P99')
    else:
        ax2.fill_between(replicas, exp_data['min_latency'], exp_data['max_latency'],
                         alpha=0.2, color='#3498db', label='Min-Max Range')
        ax2.plot(replicas, exp_data['avg_latency'], 'o-', color='#2ecc71',
                 linewidth=2.5, markersize=10, label='Average')
        ax2.plot(replicas, exp_data['p95_latency'], 's--', color='#e74c3c',
                 linewidth=2, markersize=8, label='P95')
        ax2.plot(replicas, exp_data['p99_latency'], '^:', color='#9b59b6',
                 linewidth=2, markersize=8, label='P99')

    ax2.set_xlabel('Number of Replicas')
    ax2.set_ylabel('Latency (ms)')
    ax2.set_title('(b) Latency vs. Replica Count')
    ax2.set_xticks(replicas)
    ax2.legend(loc='upper right', framealpha=0.9)

    if not publication_mode:
        for i, r in enumerate(replicas):
            ax2.annotate(f'{exp_data["avg_latency"][i]:.0f}ms',
                        xy=(r, exp_data['avg_latency'][i]),
                        xytext=(8, 0), textcoords='offset points',
                        fontsize=9, color='#2ecc71')

    # Plot 3 & 4: Resource utilization (if available)
    if has_metrics:
        # CPU usage per replica
        if publication_mode:
            ax3.fill_between(replicas, [0]*len(replicas), exp_data['max_cpu'],
                             alpha=0.15, color=COLORS['secondary'])
            ax3.plot(replicas, exp_data['avg_cpu'], 'o-', color=COLORS['secondary'],
                     linewidth=1.5, markersize=5, label='Mean CPU')
            ax3.plot(replicas, exp_data['max_cpu'], 's--', color=COLORS['secondary'],
                     linewidth=1, markersize=4, alpha=0.6, label='Max CPU')
        else:
            ax3.fill_between(replicas, [0]*len(replicas), exp_data['max_cpu'],
                             alpha=0.2, color='#e74c3c', label='Max CPU')
            ax3.plot(replicas, exp_data['avg_cpu'], 'o-', color='#e74c3c',
                     linewidth=2.5, markersize=10, label='Avg CPU per replica')

        ax3.set_xlabel('Number of Replicas')
        ax3.set_ylabel('CPU (millicores)')
        ax3.set_title('(c) Server CPU per Replica')
        ax3.set_xticks(replicas)
        ax3.legend(loc='upper right', framealpha=0.9)

        if not publication_mode:
            for i, r in enumerate(replicas):
                ax3.annotate(f'{exp_data["avg_cpu"][i]:.0f}m',
                            xy=(r, exp_data['avg_cpu'][i]),
                            xytext=(8, 0), textcoords='offset points',
                            fontsize=9, color='#e74c3c')

        # PostgreSQL CPU usage
        if has_db_metrics:
            if publication_mode:
                ax4.fill_between(replicas, [0]*len(replicas), exp_data['db_max_cpu'],
                                 alpha=0.15, color=COLORS['quaternary'])
                ax4.plot(replicas, exp_data['db_avg_cpu'], 'o-', color=COLORS['quaternary'],
                         linewidth=1.5, markersize=5, label='Mean CPU')
                ax4.plot(replicas, exp_data['db_max_cpu'], 's--', color=COLORS['quaternary'],
                         linewidth=1, markersize=4, alpha=0.6, label='Max CPU')
            else:
                ax4.fill_between(replicas, [0]*len(replicas), exp_data['db_max_cpu'],
                                 alpha=0.2, color='#9b59b6', label='Max CPU')
                ax4.plot(replicas, exp_data['db_avg_cpu'], 'o-', color='#9b59b6',
                         linewidth=2.5, markersize=10, label='Avg PostgreSQL CPU')

            ax4.set_xlabel('Number of Replicas')
            ax4.set_ylabel('CPU (millicores)')
            ax4.set_title('(d) Database CPU')
            ax4.set_xticks(replicas)
            ax4.legend(loc='upper right', framealpha=0.9)

            if not publication_mode:
                for i, r in enumerate(replicas):
                    ax4.annotate(f'{exp_data["db_avg_cpu"][i]:.0f}m',
                                xy=(r, exp_data['db_avg_cpu'][i]),
                                xytext=(8, 0), textcoords='offset points',
                                fontsize=9, color='#9b59b6')
        else:
            ax4.text(0.5, 0.5, 'No database metrics available',
                    ha='center', va='center', transform=ax4.transAxes)
            ax4.set_title('(d) Database CPU')

    plt.tight_layout()

    # Save in multiple formats for publication
    if publication_mode:
        base_name = output_file.rsplit('.', 1)[0]
        plt.savefig(f"{base_name}.pdf", format='pdf')
        plt.savefig(f"{base_name}.png", format='png')
        print(f"Saved: {base_name}.pdf")
        print(f"Saved: {base_name}.png")
    else:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Saved: {output_file}")

    plt.close()

def plot_comparison(experiments, output_file, publication_mode=False):
    """Plot multiple experiments for comparison."""
    n_exp = len(experiments)

    if publication_mode:
        setup_publication_style()
        fig, axes = plt.subplots(2, 2, figsize=(7, 5.5))
        # Use colorblind-friendly palette
        colors = [COLORS['primary'], COLORS['secondary'], COLORS['tertiary'], COLORS['quaternary']][:n_exp]
    else:
        fig, axes = plt.subplots(2, 2, figsize=(14, 10))
        fig.suptitle(f'Colonies Assign - Scaling Comparison ({n_exp} experiments)', fontsize=14, fontweight='bold')
        colors = plt.cm.tab10(np.linspace(0, 1, n_exp))

    # Plot 1: Average latency
    ax1 = axes[0, 0]
    for i, exp in enumerate(experiments):
        if publication_mode:
            label = f"{exp['executors']} executors"
        else:
            label = f"{exp['name']} ({exp['executors']}e/{exp['processes']}p)"
        ax1.plot(exp['replicas'], exp['avg_latency'], 'o-', color=colors[i],
                linewidth=1.5 if publication_mode else 2, markersize=5 if publication_mode else 8, label=label)
    ax1.set_xlabel('Number of Replicas')
    ax1.set_ylabel('Mean Latency (ms)')
    ax1.set_title('(a) Mean Latency')
    ax1.legend(fontsize=8 if publication_mode else 8, framealpha=0.9)

    # Plot 2: P95 latency
    ax2 = axes[0, 1]
    for i, exp in enumerate(experiments):
        ax2.plot(exp['replicas'], exp['p95_latency'], 's--', color=colors[i],
                linewidth=1.5 if publication_mode else 2, markersize=4 if publication_mode else 8, label=exp['name'] if not publication_mode else None)
    ax2.set_xlabel('Number of Replicas')
    ax2.set_ylabel('P95 Latency (ms)')
    ax2.set_title('(b) P95 Latency')
    if not publication_mode:
        ax2.legend(fontsize=8)

    # Plot 3: P99 latency
    ax3 = axes[1, 0]
    for i, exp in enumerate(experiments):
        ax3.plot(exp['replicas'], exp['p99_latency'], '^:', color=colors[i],
                linewidth=1.5 if publication_mode else 2, markersize=4 if publication_mode else 8, label=exp['name'] if not publication_mode else None)
    ax3.set_xlabel('Number of Replicas')
    ax3.set_ylabel('P99 Latency (ms)')
    ax3.set_title('(c) P99 Latency')
    if not publication_mode:
        ax3.legend(fontsize=8)

    # Plot 4: Speedup chart (more useful than raw latency at max replicas)
    ax4 = axes[1, 1]
    for i, exp in enumerate(experiments):
        if exp['avg_latency'] and exp['avg_latency'][0] > 0:
            baseline = exp['avg_latency'][0]
            speedup = [baseline / lat if lat > 0 else 0 for lat in exp['avg_latency']]
            ax4.plot(exp['replicas'], speedup, 'o-', color=colors[i],
                    linewidth=1.5 if publication_mode else 2, markersize=5 if publication_mode else 8,
                    label=f"{exp['executors']} executors" if publication_mode else exp['name'])

    # Add ideal linear scaling line
    if experiments:
        max_replicas = max(exp['replicas'][-1] for exp in experiments if exp['replicas'])
        ideal_x = list(range(1, max_replicas + 1))
        ideal_y = ideal_x
        ax4.plot(ideal_x, ideal_y, 'k--', alpha=0.5, linewidth=1, label='Ideal (linear)')

    ax4.set_xlabel('Number of Replicas')
    ax4.set_ylabel('Speedup (vs. 1 replica)')
    ax4.set_title('(d) Scaling Efficiency')
    ax4.legend(fontsize=8 if publication_mode else 8, framealpha=0.9)

    plt.tight_layout()

    if publication_mode:
        base_name = output_file.rsplit('.', 1)[0]
        plt.savefig(f"{base_name}.pdf", format='pdf')
        plt.savefig(f"{base_name}.png", format='png')
        print(f"Saved: {base_name}.pdf")
        print(f"Saved: {base_name}.png")
    else:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Saved: {output_file}")

    plt.close()

def plot_replica_timeseries(exp_dir, replica_count, output_file, publication_mode=False):
    """Plot CPU load over time for each Colonies replica."""
    metrics_file = os.path.join(exp_dir, f"replicas_{replica_count}", "pod_metrics_timeseries.csv")

    if not os.path.exists(metrics_file):
        print(f"No timeseries metrics found: {metrics_file}")
        return

    # Read timeseries data
    data = {}  # pod -> [(timestamp, cpu, memory), ...]
    with open(metrics_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            pod = row['pod']
            if 'colonies-server' in pod:
                if pod not in data:
                    data[pod] = []
                data[pod].append((
                    int(row['timestamp']),
                    float(row['cpu_millicores']),
                    float(row['memory_mib'])
                ))

    if not data:
        print("No colonies-server metrics found")
        return

    # Also get PostgreSQL data
    postgres_data = []
    with open(metrics_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            if 'postgres' in row['pod']:
                postgres_data.append((
                    int(row['timestamp']),
                    float(row['cpu_millicores']),
                    float(row['memory_mib'])
                ))

    if publication_mode:
        setup_publication_style()
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(7, 5), sharex=True)
    else:
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 8), sharex=True)
        fig.suptitle(f'CPU Load Over Time - {replica_count} Replicas', fontsize=14, fontweight='bold')

    # Color palette for replicas
    colors = plt.cm.tab10(np.linspace(0, 1, len(data)))

    # Plot each replica's CPU
    for i, (pod, points) in enumerate(sorted(data.items())):
        if not points:
            continue
        # Normalize timestamps to start from 0
        start_time = min(p[0] for p in points)
        times = [(p[0] - start_time) for p in points]
        cpus = [p[1] for p in points]

        # Extract replica number from pod name (e.g., colonies-server-0 -> 0)
        try:
            replica_num = int(pod.split('-')[-1])
            label = f'Replica {replica_num}'
        except:
            label = pod

        ax1.plot(times, cpus, '-', color=colors[i], linewidth=1.5, label=label, alpha=0.8)

    ax1.set_ylabel('CPU (millicores)')
    ax1.set_title('(a) Colonies Server Replicas CPU')
    ax1.legend(loc='upper right', fontsize=8, ncol=2)
    ax1.grid(True, alpha=0.3)

    # Plot PostgreSQL CPU
    if postgres_data:
        start_time = min(p[0] for p in postgres_data)
        times = [(p[0] - start_time) for p in postgres_data]
        cpus = [p[1] for p in postgres_data]
        ax2.plot(times, cpus, '-', color=COLORS['quaternary'], linewidth=1.5, label='PostgreSQL')
        ax2.fill_between(times, 0, cpus, alpha=0.2, color=COLORS['quaternary'])

    ax2.set_xlabel('Time (seconds)')
    ax2.set_ylabel('CPU (millicores)')
    ax2.set_title('(b) PostgreSQL CPU')
    ax2.legend(loc='upper right', fontsize=8)
    ax2.grid(True, alpha=0.3)

    plt.tight_layout()

    if publication_mode:
        base_name = output_file.rsplit('.', 1)[0]
        plt.savefig(f"{base_name}.pdf", format='pdf')
        plt.savefig(f"{base_name}.png", format='png')
        print(f"Saved: {base_name}.pdf")
        print(f"Saved: {base_name}.png")
    else:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Saved: {output_file}")

    plt.close()


def plot_cpu_with_logs(exp_dir, replica_count, output_file, publication_mode=False):
    """Plot CPU utilization with log event density overlay."""
    import re
    from collections import defaultdict

    metrics_file = os.path.join(exp_dir, f"replicas_{replica_count}", "pod_metrics_timeseries.csv")
    logs_file = os.path.join(exp_dir, f"replicas_{replica_count}", "colonies_logs.jsonl")

    # Check parent dir for logs
    if not os.path.exists(logs_file):
        logs_file = os.path.join(exp_dir, "colonies_logs.jsonl")

    if not os.path.exists(metrics_file):
        print(f"No metrics found: {metrics_file}")
        return

    # Read CPU metrics
    cpu_by_second = defaultdict(float)
    with open(metrics_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            if 'colonies-server' in row['pod']:
                ts = int(row['timestamp'])
                cpu_by_second[ts] += float(row['cpu_millicores'])

    if not cpu_by_second:
        print("No CPU data found")
        return

    # Read log density if available
    log_counts = defaultdict(int)
    has_logs = False
    if os.path.exists(logs_file):
        timestamp_re = re.compile(r'^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})')
        with open(logs_file, 'r') as f:
            for line in f:
                match = timestamp_re.match(line)
                if match:
                    try:
                        from datetime import datetime
                        dt = datetime.fromisoformat(match.group(1) + '+00:00')
                        ts = int(dt.timestamp())
                        log_counts[ts] += 1
                        has_logs = True
                    except:
                        pass

    # Prepare data
    start_time = min(cpu_by_second.keys())
    times = sorted(cpu_by_second.keys())
    rel_times = [(t - start_time) for t in times]
    cpu_values = [cpu_by_second[t] for t in times]

    if publication_mode:
        setup_publication_style()
        fig, ax1 = plt.subplots(figsize=(7, 4))
    else:
        fig, ax1 = plt.subplots(figsize=(12, 5))
        fig.suptitle(f'CPU Utilization with Log Events - {replica_count} Replicas', fontsize=14, fontweight='bold')

    # Plot CPU
    ax1.plot(rel_times, cpu_values, '-', color=COLORS['primary'], linewidth=1.5, label='Total CPU')
    ax1.fill_between(rel_times, 0, cpu_values, alpha=0.2, color=COLORS['primary'])
    ax1.set_xlabel('Time (seconds)')
    ax1.set_ylabel('CPU (millicores)', color=COLORS['primary'])
    ax1.tick_params(axis='y', labelcolor=COLORS['primary'])
    ax1.grid(True, alpha=0.3)

    # Plot log density on secondary axis
    if has_logs:
        ax2 = ax1.twinx()
        log_times = [(t - start_time) for t in sorted(log_counts.keys()) if t in cpu_by_second or any(abs(t - ct) <= 1 for ct in cpu_by_second)]
        log_vals = [log_counts[t + start_time] for t in log_times]

        if log_times:
            ax2.bar(log_times, log_vals, alpha=0.3, color=COLORS['secondary'], width=0.8, label='Log Events/sec')
            ax2.set_ylabel('Log Events per Second', color=COLORS['secondary'])
            ax2.tick_params(axis='y', labelcolor=COLORS['secondary'])

    # Legend
    lines1, labels1 = ax1.get_legend_handles_labels()
    if has_logs:
        lines2, labels2 = ax2.get_legend_handles_labels()
        ax1.legend(lines1 + lines2, labels1 + labels2, loc='upper right')
    else:
        ax1.legend(loc='upper right')

    plt.tight_layout()

    if publication_mode:
        base_name = output_file.rsplit('.', 1)[0]
        plt.savefig(f"{base_name}.pdf", format='pdf')
        plt.savefig(f"{base_name}.png", format='png')
        print(f"Saved: {base_name}.pdf")
    else:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Saved: {output_file}")

    plt.close()


def plot_app_metrics(exp_dir, replica_count, output_file, publication_mode=False):
    """Plot Colonies application metrics over time (processes waiting/running/completed)."""
    metrics_file = os.path.join(exp_dir, f"replicas_{replica_count}", "app_metrics.csv")

    if not os.path.exists(metrics_file):
        print(f"No app metrics found: {metrics_file}")
        return

    # Read app metrics data
    timestamps = []
    waiting = []
    running = []
    successful = []
    failed = []

    with open(metrics_file, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            timestamps.append(int(row['timestamp']))
            waiting.append(int(row['processes_waiting']))
            running.append(int(row['processes_running']))
            successful.append(int(row['processes_successful']))
            failed.append(int(row['processes_failed']))

    if not timestamps:
        print("No app metrics data found")
        return

    # Normalize timestamps to start from 0
    start_time = min(timestamps)
    times = [(t - start_time) for t in timestamps]

    if publication_mode:
        setup_publication_style()
        fig, ax = plt.subplots(figsize=(7, 4))
    else:
        fig, ax = plt.subplots(figsize=(12, 6))
        fig.suptitle(f'Process States Over Time - {replica_count} Replicas', fontsize=14, fontweight='bold')

    # Plot stacked areas
    ax.fill_between(times, 0, waiting, alpha=0.5, label='Waiting', color=COLORS['primary'])
    ax.fill_between(times, waiting, [w + r for w, r in zip(waiting, running)], alpha=0.5, label='Running', color=COLORS['secondary'])
    ax.fill_between(times, [w + r for w, r in zip(waiting, running)],
                    [w + r + s for w, r, s in zip(waiting, running, successful)], alpha=0.5, label='Successful', color=COLORS['tertiary'])

    # Also plot lines for clarity
    ax.plot(times, waiting, '-', linewidth=1.5, color=COLORS['primary'])
    ax.plot(times, [w + r for w, r in zip(waiting, running)], '-', linewidth=1.5, color=COLORS['secondary'])
    ax.plot(times, [w + r + s for w, r, s in zip(waiting, running, successful)], '-', linewidth=1.5, color=COLORS['tertiary'])

    ax.set_xlabel('Time (seconds)')
    ax.set_ylabel('Process Count')
    ax.set_title('Process States Over Time')
    ax.legend(loc='upper right', fontsize=9)
    ax.grid(True, alpha=0.3)

    plt.tight_layout()

    if publication_mode:
        base_name = output_file.rsplit('.', 1)[0]
        plt.savefig(f"{base_name}.pdf", format='pdf')
        plt.savefig(f"{base_name}.png", format='png')
        print(f"Saved: {base_name}.pdf")
        print(f"Saved: {base_name}.png")
    else:
        plt.savefig(output_file, dpi=150, bbox_inches='tight')
        print(f"Saved: {output_file}")

    plt.close()


def main():
    parser = argparse.ArgumentParser(description='Plot Colonies scaling benchmark results')
    parser.add_argument('experiments', nargs='*', help='Experiment directories to plot')
    parser.add_argument('--latest', action='store_true', help='Plot only the latest experiment')
    parser.add_argument('--list', action='store_true', help='List available experiments')
    parser.add_argument('--output', '-o', help='Output directory (default: experiment dir)')
    parser.add_argument('--publication', '-p', action='store_true', help='Publication mode: cleaner styling, no timestamps')
    parser.add_argument('--timeseries', '-t', type=int, metavar='REPLICAS', help='Plot CPU timeseries for specific replica count')
    parser.add_argument('--app-metrics', '-a', type=int, metavar='REPLICAS', help='Plot app metrics timeseries for specific replica count')
    parser.add_argument('--cpu-logs', '-l', type=int, metavar='REPLICAS', help='Plot CPU with log event density overlay')
    args = parser.parse_args()

    all_experiments = find_all_experiments()

    if args.list:
        print("Available experiments:")
        for exp in all_experiments:
            name = parse_experiment_name(exp)
            data = read_experiment(exp)
            print(f"  {name}: {len(data['replicas'])} replicas, {data['executors']} executors, {data['processes']} processes")
            print(f"    Path: {exp}")
        return

    if not all_experiments:
        print("No experiments found in results/")
        sys.exit(1)

    # Determine which experiments to plot
    if args.experiments:
        exp_dirs = args.experiments
    elif args.latest:
        exp_dirs = [all_experiments[-1]]
    else:
        exp_dirs = all_experiments

    # Handle timeseries plotting
    if args.timeseries is not None:
        for exp_dir in exp_dirs:
            if os.path.isdir(exp_dir):
                out_dir = args.output or exp_dir
                filename = f"cpu_timeseries_r{args.timeseries}_publication.png" if args.publication else f"cpu_timeseries_r{args.timeseries}.png"
                plot_replica_timeseries(exp_dir, args.timeseries, os.path.join(out_dir, filename), publication_mode=args.publication)
        return

    # Handle app metrics plotting
    if args.app_metrics is not None:
        for exp_dir in exp_dirs:
            if os.path.isdir(exp_dir):
                out_dir = args.output or exp_dir
                filename = f"app_metrics_r{args.app_metrics}_publication.png" if args.publication else f"app_metrics_r{args.app_metrics}.png"
                plot_app_metrics(exp_dir, args.app_metrics, os.path.join(out_dir, filename), publication_mode=args.publication)
        return

    # Handle CPU + logs plotting
    if args.cpu_logs is not None:
        for exp_dir in exp_dirs:
            if os.path.isdir(exp_dir):
                out_dir = args.output or exp_dir
                filename = f"cpu_logs_r{args.cpu_logs}_publication.png" if args.publication else f"cpu_logs_r{args.cpu_logs}.png"
                plot_cpu_with_logs(exp_dir, args.cpu_logs, os.path.join(out_dir, filename), publication_mode=args.publication)
        return

    # Read experiment data
    experiments = []
    for exp_dir in exp_dirs:
        if os.path.isdir(exp_dir):
            data = read_experiment(exp_dir)
            if data['replicas']:
                experiments.append(data)
            else:
                print(f"Warning: No results found in {exp_dir}")

    if not experiments:
        print("No valid experiments to plot")
        sys.exit(1)

    # Generate one plot per experiment
    for exp in experiments:
        out_dir = args.output or exp['path']
        filename = "scaling_results_publication.png" if args.publication else "scaling_results.png"
        plot_experiment(exp, os.path.join(out_dir, filename), publication_mode=args.publication)

        # Also generate timeseries for each replica count
        for replica_count in exp['replicas']:
            ts_filename = f"cpu_timeseries_r{replica_count}.png"
            plot_replica_timeseries(exp['path'], replica_count, os.path.join(out_dir, ts_filename), publication_mode=False)

            # Generate app metrics plot if data exists
            app_filename = f"app_metrics_r{replica_count}.png"
            plot_app_metrics(exp['path'], replica_count, os.path.join(out_dir, app_filename), publication_mode=False)

            # Generate CPU + logs plot if logs exist
            cpu_logs_filename = f"cpu_logs_r{replica_count}.png"
            plot_cpu_with_logs(exp['path'], replica_count, os.path.join(out_dir, cpu_logs_filename), publication_mode=False)

    # If multiple experiments, also generate comparison
    if len(experiments) > 1:
        out_dir = args.output or os.path.join(get_script_dir(), "results")
        filename = "experiments_comparison_publication.png" if args.publication else "experiments_comparison.png"
        plot_comparison(experiments, os.path.join(out_dir, filename), publication_mode=args.publication)

    print(f"\nGenerated {len(experiments)} plot(s)")

if __name__ == "__main__":
    main()
