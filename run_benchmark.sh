#!/bin/bash

echo "======================================"
echo "Matrix GC Benchmark Comparison"
echo "Standard GC vs Green Tea GC"
echo "======================================"
echo ""

# Check Go version
GO_VERSION=$(go version)
echo "Go Version: $GO_VERSION"
echo ""

# Create output directory
mkdir -p benchmark_results

# Run with standard GC
echo "======================================"
echo "Running with STANDARD GC..."
echo "======================================"
echo ""

go build -o matrix_benchmark_standard matrix_gc_benchmark.go
if [ $? -ne 0 ]; then
    echo "Build failed for standard GC"
    exit 1
fi

echo "Running benchmark (this may take a minute)..."
./matrix_benchmark_standard | tee benchmark_results/standard_gc.txt
echo ""

# Run with Green Tea GC
echo "======================================"
echo "Running with GREEN TEA GC..."
echo "======================================"
echo ""

GOEXPERIMENT=greenteagc go build -o matrix_benchmark_greentea matrix_gc_benchmark.go
if [ $? -ne 0 ]; then
    echo "Build failed for Green Tea GC"
    echo "Note: Green Tea GC is only available in Go 1.25+"
    exit 1
fi

echo "Running benchmark (this may take a minute)..."
./matrix_benchmark_greentea | tee benchmark_results/greentea_gc.txt
echo ""

# Extract and compare key metrics
echo "======================================"
echo "COMPARISON SUMMARY"
echo "======================================"
echo ""

extract_metric() {
    grep "$2" "$1" | awk -F': ' '{print $2}' | awk '{print $1}'
}

STANDARD_DURATION=$(extract_metric "benchmark_results/standard_gc.txt" "Total Duration")
GREENTEA_DURATION=$(extract_metric "benchmark_results/greentea_gc.txt" "Total Duration")

STANDARD_GC_COUNT=$(extract_metric "benchmark_results/standard_gc.txt" "Number of GCs")
GREENTEA_GC_COUNT=$(extract_metric "benchmark_results/greentea_gc.txt" "Number of GCs")

STANDARD_GC_PAUSE=$(extract_metric "benchmark_results/standard_gc.txt" "Total GC Pause")
GREENTEA_GC_PAUSE=$(extract_metric "benchmark_results/greentea_gc.txt" "Total GC Pause")

STANDARD_GC_CPU=$(extract_metric "benchmark_results/standard_gc.txt" "GC CPU Fraction")
GREENTEA_GC_CPU=$(extract_metric "benchmark_results/greentea_gc.txt" "GC CPU Fraction")

echo "Metric                    | Standard GC      | Green Tea GC     | Improvement"
echo "--------------------------|------------------|------------------|-------------"
printf "Total Duration            | %-16s | %-16s | " "$STANDARD_DURATION" "$GREENTEA_DURATION"
if [ ! -z "$STANDARD_DURATION" ] && [ ! -z "$GREENTEA_DURATION" ]; then
    # Calculate percentage improvement (bash doesn't do floating point well, so we'll just show the values)
    echo "See raw data"
else
    echo "N/A"
fi

printf "Number of GCs             | %-16s | %-16s | " "$STANDARD_GC_COUNT" "$GREENTEA_GC_COUNT"
echo ""

printf "Total GC Pause            | %-16s | %-16s | " "$STANDARD_GC_PAUSE" "$GREENTEA_GC_PAUSE"
echo ""

printf "GC CPU Fraction           | %-16s | %-16s | " "$STANDARD_GC_CPU%" "$GREENTEA_GC_CPU%"
echo ""

echo ""
echo "Full results saved to:"
echo "  - benchmark_results/standard_gc.txt"
echo "  - benchmark_results/greentea_gc.txt"
echo ""

# Cleanup
rm -f matrix_benchmark_standard matrix_benchmark_greentea

echo "Benchmark complete!"