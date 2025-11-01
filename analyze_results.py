#!/usr/bin/env python3
"""
Analyze and compare GC benchmark results
"""

import re
import sys
from pathlib import Path

def parse_duration(duration_str):
    """Convert duration string to milliseconds"""
    if 's' in duration_str:
        return float(duration_str.replace('s', '')) * 1000
    elif 'ms' in duration_str:
        return float(duration_str.replace('ms', ''))
    elif 'µs' in duration_str or 'us' in duration_str:
        return float(duration_str.replace('µs', '').replace('us', '')) / 1000
    return 0

def extract_metrics(filename):
    """Extract key metrics from benchmark output"""
    try:
        with open(filename, 'r') as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: {filename} not found")
        return None
    
    metrics = {}
    
    # Extract metrics using regex
    patterns = {
        'duration': r'Total Duration:\s*([\d.]+[µms]+)',
        'ops_per_sec': r'Operations/sec:\s*([\d.]+)',
        'total_alloc': r'Total Allocated:\s*([\d.]+)\s*MB',
        'heap_alloc': r'Heap Allocated:\s*([\d.]+)\s*MB',
        'heap_objects': r'Heap Objects:\s*(\d+)',
        'num_gc': r'Number of GCs:\s*(\d+)',
        'total_pause': r'Total GC Pause:\s*([\d.]+[µms]+)',
        'avg_pause': r'Average GC Pause:\s*([\d.]+[µms]+)',
        'gc_pause_overhead': r'GC Pause Overhead:\s*([\d.]+)%',
        'gc_cpu_fraction': r'GC CPU Fraction:\s*([\d.]+)%',
        'time_per_iter': r'Time per iteration:\s*([\d.]+[µms]+)',
    }
    
    for key, pattern in patterns.items():
        match = re.search(pattern, content)
        if match:
            metrics[key] = match.group(1)
    
    return metrics

def calculate_improvement(standard, greentea):
    """Calculate percentage improvement"""
    try:
        std = float(standard.replace('%', '').replace('ms', '').replace('s', '').replace('MB', '').replace('µs', ''))
        gt = float(greentea.replace('%', '').replace('ms', '').replace('s', '').replace('MB', '').replace('µs', ''))
        improvement = ((std - gt) / std) * 100
        return improvement
    except:
        return None

def main():
    results_dir = Path("benchmark_results")
    
    if not results_dir.exists():
        print("Error: benchmark_results directory not found")
        print("Please run ./run_benchmark.sh first")
        sys.exit(1)
    
    standard_file = results_dir / "standard_gc.txt"
    greentea_file = results_dir / "greentea_gc.txt"
    
    print("=" * 80)
    print("DETAILED GC BENCHMARK COMPARISON")
    print("=" * 80)
    print()
    
    standard_metrics = extract_metrics(standard_file)
    greentea_metrics = extract_metrics(greentea_file)
    
    if not standard_metrics or not greentea_metrics:
        print("Error: Could not parse benchmark results")
        sys.exit(1)
    
    # Print comparison table
    print(f"{'Metric':<30} | {'Standard GC':<20} | {'Green Tea GC':<20} | {'Change':<15}")
    print("-" * 95)
    
    metrics_to_compare = [
        ('Total Duration', 'duration', 'lower'),
        ('Operations/sec', 'ops_per_sec', 'higher'),
        ('Total Memory Allocated', 'total_alloc', 'neutral'),
        ('Heap Allocated', 'heap_alloc', 'neutral'),
        ('Heap Objects', 'heap_objects', 'neutral'),
        ('Number of GCs', 'num_gc', 'lower'),
        ('Total GC Pause', 'total_pause', 'lower'),
        ('Average GC Pause', 'avg_pause', 'lower'),
        ('GC Pause Overhead', 'gc_pause_overhead', 'lower'),
        ('GC CPU Fraction', 'gc_cpu_fraction', 'lower'),
        ('Time per Iteration', 'time_per_iter', 'lower'),
    ]
    
    improvements = []
    
    for display_name, key, direction in metrics_to_compare:
        if key in standard_metrics and key in greentea_metrics:
            std_val = standard_metrics[key]
            gt_val = greentea_metrics[key]
            
            improvement = calculate_improvement(std_val, gt_val)
            
            if improvement is not None:
                if direction == 'lower':
                    change_str = f"{improvement:+.2f}%"
                    if improvement > 0:
                        change_str += " ✓"
                elif direction == 'higher':
                    change_str = f"{-improvement:+.2f}%"
                    if improvement < 0:
                        change_str += " ✓"
                else:
                    change_str = f"{improvement:+.2f}%"
                
                improvements.append((display_name, improvement, direction))
            else:
                change_str = "N/A"
            
            print(f"{display_name:<30} | {std_val:<20} | {gt_val:<20} | {change_str:<15}")
    
    print()
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)
    print()
    
    # Calculate overall GC improvement
    if 'gc_cpu_fraction' in standard_metrics and 'gc_cpu_fraction' in greentea_metrics:
        gc_improvement = calculate_improvement(
            standard_metrics['gc_cpu_fraction'],
            greentea_metrics['gc_cpu_fraction']
        )
        if gc_improvement:
            print(f"🎯 GC CPU Time Reduction: {gc_improvement:.2f}%")
    
    if 'duration' in standard_metrics and 'duration' in greentea_metrics:
        duration_improvement = calculate_improvement(
            standard_metrics['duration'],
            greentea_metrics['duration']
        )
        if duration_improvement:
            print(f"⚡ Overall Performance Improvement: {duration_improvement:.2f}%")
    
    if 'total_pause' in standard_metrics and 'total_pause' in greentea_metrics:
        pause_improvement = calculate_improvement(
            standard_metrics['total_pause'],
            greentea_metrics['total_pause']
        )
        if pause_improvement:
            print(f"⏸️  GC Pause Reduction: {pause_improvement:.2f}%")
    
    print()
    print("Legend:")
    print("  ✓ = Improved in desired direction")
    print("  Lower is better for: Duration, GC metrics, Pause times")
    print("  Higher is better for: Operations/sec")
    print()
    
    # Analysis
    print("=" * 80)
    print("ANALYSIS")
    print("=" * 80)
    print()
    
    if gc_improvement and gc_improvement > 0:
        if gc_improvement >= 30:
            print("🌟 Excellent! Green Tea shows significant GC improvement (>30%)")
        elif gc_improvement >= 15:
            print("✅ Good! Green Tea shows solid GC improvement (15-30%)")
        elif gc_improvement >= 5:
            print("👍 Moderate improvement from Green Tea (5-15%)")
        else:
            print("📊 Small improvement from Green Tea (<5%)")
        print()
        print("This workload benefits from Green Tea's page-based scanning approach.")
        print("The regular object sizes and allocation patterns allow effective page accumulation.")
    else:
        print("⚠️  Limited or no improvement observed.")
        print("This workload may not benefit significantly from Green Tea.")
        print("Consider:")
        print("  - Increasing matrix size for more objects per page")
        print("  - Different allocation patterns")
        print("  - Your specific workload characteristics")
    
    print()

if __name__ == "__main__":
    main()