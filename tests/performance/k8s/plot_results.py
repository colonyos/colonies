#!/usr/bin/env python3
"""
Plot scaling comparison results from Colonies performance benchmarks.

Usage:
    python3 plot_results.py                    # Plot all experiments
    python3 plot_results.py --latest           # Plot only the latest
    python3 plot_results.py --list             # List available experiments
    python3 plot_results.py <dir1> <dir2> ...  # Plot specific experiments
"""

import sys
import os
import glob
import csv
import argparse
import matplotlib.pyplot as plt
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
        'executors': 0,
        'processes': 0,
    }

    for r in range(1, 10):
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
            if results['executors'] == 0:
                results['executors'] = int(data.get('executors', 0))
                results['processes'] = int(data.get('processes', 0))

    return results

def plot_experiment(exp_data, output_file):
    """Plot single combined chart for an experiment."""
    replicas = exp_data['replicas']

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 6))
    fig.suptitle(f"Colonies Assign Latency - {exp_data['name']}\n({exp_data['executors']} executors, {exp_data['processes']} processes)",
                 fontsize=14, fontweight='bold')

    # Left plot: Bar chart
    x = np.arange(len(replicas))
    width = 0.18

    bars1 = ax1.bar(x - width*1.5, exp_data['avg_latency'], width, label='Average', color='#2ecc71')
    bars2 = ax1.bar(x - width*0.5, exp_data['p50_latency'], width, label='P50', color='#3498db')
    bars3 = ax1.bar(x + width*0.5, exp_data['p95_latency'], width, label='P95', color='#e74c3c')
    bars4 = ax1.bar(x + width*1.5, exp_data['p99_latency'], width, label='P99', color='#9b59b6')

    ax1.set_xlabel('Number of Replicas', fontsize=12)
    ax1.set_ylabel('Latency (ms)', fontsize=12)
    ax1.set_title('Latency Percentiles')
    ax1.set_xticks(x)
    ax1.set_xticklabels([str(r) for r in replicas])
    ax1.legend(loc='upper right')
    ax1.grid(axis='y', alpha=0.3)

    for bars in [bars1, bars2, bars3, bars4]:
        for bar in bars:
            height = bar.get_height()
            ax1.annotate(f'{height:.0f}',
                        xy=(bar.get_x() + bar.get_width() / 2, height),
                        xytext=(0, 3),
                        textcoords="offset points",
                        ha='center', va='bottom', fontsize=8)

    # Right plot: Line chart
    ax2.fill_between(replicas, exp_data['min_latency'], exp_data['max_latency'],
                     alpha=0.2, color='#3498db', label='Min-Max Range')
    ax2.plot(replicas, exp_data['avg_latency'], 'o-', color='#2ecc71',
             linewidth=2.5, markersize=10, label='Average')
    ax2.plot(replicas, exp_data['p95_latency'], 's--', color='#e74c3c',
             linewidth=2, markersize=8, label='P95')
    ax2.plot(replicas, exp_data['p99_latency'], '^:', color='#9b59b6',
             linewidth=2, markersize=8, label='P99')

    ax2.set_xlabel('Number of Replicas', fontsize=12)
    ax2.set_ylabel('Latency (ms)', fontsize=12)
    ax2.set_title('Latency Distribution')
    ax2.set_xticks(replicas)
    ax2.legend(loc='upper right')
    ax2.grid(alpha=0.3)

    for i, r in enumerate(replicas):
        ax2.annotate(f'{exp_data["avg_latency"][i]:.0f}ms',
                    xy=(r, exp_data['avg_latency'][i]),
                    xytext=(8, 0), textcoords='offset points',
                    fontsize=9, color='#2ecc71')

    plt.tight_layout()
    plt.savefig(output_file, dpi=150, bbox_inches='tight')
    plt.close()
    print(f"Saved: {output_file}")

def plot_comparison(experiments, output_file):
    """Plot multiple experiments for comparison."""
    n_exp = len(experiments)

    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    fig.suptitle(f'Colonies Assign - Scaling Comparison ({n_exp} experiments)', fontsize=14, fontweight='bold')

    colors = plt.cm.tab10(np.linspace(0, 1, n_exp))

    # Plot 1: Average latency
    ax1 = axes[0, 0]
    for i, exp in enumerate(experiments):
        label = f"{exp['name']} ({exp['executors']}e/{exp['processes']}p)"
        ax1.plot(exp['replicas'], exp['avg_latency'], 'o-', color=colors[i],
                linewidth=2, markersize=8, label=label)
    ax1.set_xlabel('Replicas')
    ax1.set_ylabel('Avg Latency (ms)')
    ax1.set_title('Average Latency')
    ax1.legend(fontsize=8)
    ax1.grid(alpha=0.3)

    # Plot 2: P95 latency
    ax2 = axes[0, 1]
    for i, exp in enumerate(experiments):
        ax2.plot(exp['replicas'], exp['p95_latency'], 's--', color=colors[i],
                linewidth=2, markersize=8, label=exp['name'])
    ax2.set_xlabel('Replicas')
    ax2.set_ylabel('P95 Latency (ms)')
    ax2.set_title('P95 Latency')
    ax2.legend(fontsize=8)
    ax2.grid(alpha=0.3)

    # Plot 3: P99 latency
    ax3 = axes[1, 0]
    for i, exp in enumerate(experiments):
        ax3.plot(exp['replicas'], exp['p99_latency'], '^:', color=colors[i],
                linewidth=2, markersize=8, label=exp['name'])
    ax3.set_xlabel('Replicas')
    ax3.set_ylabel('P99 Latency (ms)')
    ax3.set_title('P99 Latency')
    ax3.legend(fontsize=8)
    ax3.grid(alpha=0.3)

    # Plot 4: Summary bar chart
    ax4 = axes[1, 1]
    max_replica_avg = [exp['avg_latency'][-1] if exp['avg_latency'] else 0 for exp in experiments]
    max_replicas = [exp['replicas'][-1] if exp['replicas'] else 0 for exp in experiments]

    x = np.arange(len(experiments))
    bars = ax4.bar(x, max_replica_avg, color=colors)
    ax4.set_xlabel('Experiment')
    ax4.set_ylabel('Avg Latency (ms)')
    ax4.set_title('Avg Latency at Max Replicas')
    ax4.set_xticks(x)
    ax4.set_xticklabels([f"{exp['name']}\n({r}r)" for exp, r in zip(experiments, max_replicas)],
                        fontsize=8, rotation=45, ha='right')
    ax4.grid(axis='y', alpha=0.3)

    for bar, val in zip(bars, max_replica_avg):
        ax4.annotate(f'{val:.0f}', xy=(bar.get_x() + bar.get_width()/2, val),
                    xytext=(0, 3), textcoords="offset points", ha='center', fontsize=9)

    plt.tight_layout()
    plt.savefig(output_file, dpi=150, bbox_inches='tight')
    plt.close()
    print(f"Saved: {output_file}")

def main():
    parser = argparse.ArgumentParser(description='Plot Colonies scaling benchmark results')
    parser.add_argument('experiments', nargs='*', help='Experiment directories to plot')
    parser.add_argument('--latest', action='store_true', help='Plot only the latest experiment')
    parser.add_argument('--list', action='store_true', help='List available experiments')
    parser.add_argument('--output', '-o', help='Output directory (default: experiment dir)')
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
        plot_experiment(exp, os.path.join(out_dir, "scaling_results.png"))

    # If multiple experiments, also generate comparison
    if len(experiments) > 1:
        out_dir = args.output or os.path.join(get_script_dir(), "results")
        plot_comparison(experiments, os.path.join(out_dir, "experiments_comparison.png"))

    print(f"\nGenerated {len(experiments)} plot(s)")

if __name__ == "__main__":
    main()
