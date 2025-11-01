# Matrix GC Benchmark: Standard GC vs Green Tea GC

This benchmark demonstrates the performance differences between Go's standard garbage collector and the new **Green Tea GC** introduced in Go 1.25.

## Overview

According to the Go blog post on Green Tea GC, the new garbage collector can reduce GC CPU costs by **10-40%** for many workloads. The key innovation is that Green Tea works with **pages instead of objects**, which:

1. Reduces memory access patterns that cause CPU cache misses
2. Enables better use of vector hardware (AVX-512)
3. Improves parallelization by reducing work queue contention

## License

This benchmark is provided as-is for educational and testing purposes.