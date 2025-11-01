package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"time"
)

// Matrix represents a 2D matrix with heap-allocated rows
type Matrix struct {
	rows int
	cols int
	data [][]*float64 // Slice of slices of pointers - creates lots of heap objects
}

// NewMatrix creates a new matrix with the given dimensions
func NewMatrix(rows, cols int) *Matrix {
	m := &Matrix{
		rows: rows,
		cols: cols,
		data: make([][]*float64, rows),
	}
	
	for i := 0; i < rows; i++ {
		m.data[i] = make([]*float64, cols)
		for j := 0; j < cols; j++ {
			val := rand.Float64()
			m.data[i][j] = &val // Each element is a pointer to heap-allocated float64
		}
	}
	
	return m
}

// Multiply performs matrix multiplication
func (m *Matrix) Multiply(other *Matrix) *Matrix {
	if m.cols != other.rows {
		panic("incompatible dimensions for multiplication")
	}
	
	result := NewMatrix(m.rows, other.cols)
	
	for i := 0; i < m.rows; i++ {
		for j := 0; j < other.cols; j++ {
			sum := 0.0
			for k := 0; k < m.cols; k++ {
				sum += *m.data[i][k] * *other.data[k][j]
			}
			*result.data[i][j] = sum
		}
	}
	
	return result
}

// Add performs matrix addition
func (m *Matrix) Add(other *Matrix) *Matrix {
	if m.rows != other.rows || m.cols != other.cols {
		panic("incompatible dimensions for addition")
	}
	
	result := NewMatrix(m.rows, m.cols)
	
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			*result.data[i][j] = *m.data[i][j] + *other.data[i][j]
		}
	}
	
	return result
}

// Transpose creates a transposed version of the matrix
func (m *Matrix) Transpose() *Matrix {
	result := NewMatrix(m.cols, m.rows)
	
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			*result.data[j][i] = *m.data[i][j]
		}
	}
	
	return result
}

// ScalarMultiply multiplies each element by a scalar
func (m *Matrix) ScalarMultiply(scalar float64) *Matrix {
	result := NewMatrix(m.rows, m.cols)
	
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			*result.data[i][j] = *m.data[i][j] * scalar
		}
	}
	
	return result
}

// GCStats holds garbage collection statistics
type GCStats struct {
	NumGC      uint32
	PauseTotal time.Duration
	LastPause  time.Duration
}

func getGCStats() GCStats {
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	
	lastPause := time.Duration(0)
	if len(stats.Pause) > 0 {
		lastPause = stats.Pause[0]
	}
	
	return GCStats{
		NumGC:      uint32(stats.NumGC),
		PauseTotal: stats.PauseTotal,
		LastPause:  lastPause,
	}
}

func main() {
	fmt.Println("=== Matrix GC Benchmark ===")
	fmt.Println("Comparing GC performance with heavy heap allocation")
	fmt.Println()
	
	// Get GC info
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	fmt.Println()
	
	// Configuration
	const (
		matrixSize  = 50   // Size of matrices
		iterations  = 1000 // Number of iterations
		warmupIters = 100  // Warmup iterations
	)
	
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Matrix Size: %dx%d\n", matrixSize, matrixSize)
	fmt.Printf("  Iterations: %d (+ %d warmup)\n", iterations, warmupIters)
	fmt.Println()
	
	// Warmup phase
	fmt.Println("Running warmup...")
	for i := 0; i < warmupIters; i++ {
		m1 := NewMatrix(matrixSize, matrixSize)
		m2 := NewMatrix(matrixSize, matrixSize)
		_ = m1.Multiply(m2)
	}
	
	// Force GC before benchmark
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	
	// Capture initial GC stats
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	gcStatsBefore := getGCStats()
	
	fmt.Println("Starting benchmark...")
	startTime := time.Now()
	
	// Main benchmark loop
	var results []*Matrix
	for i := 0; i < iterations; i++ {
		// Create matrices
		m1 := NewMatrix(matrixSize, matrixSize)
		m2 := NewMatrix(matrixSize, matrixSize)
		
		// Perform operations (creates many intermediate objects)
		m3 := m1.Multiply(m2)
		m4 := m1.Add(m2)
		m5 := m3.Transpose()
		m6 := m4.ScalarMultiply(2.5)
		m7 := m5.Add(m6)
		
		// Keep some results to prevent optimization away
		if i%100 == 0 {
			results = append(results, m7)
		}
	}
	
	duration := time.Since(startTime)
	
	// Capture final GC stats
	runtime.GC() // Force final GC to get accurate stats
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)
	gcStatsAfter := getGCStats()
	
	// Calculate differences
	numGCs := gcStatsAfter.NumGC - gcStatsBefore.NumGC
	totalPause := gcStatsAfter.PauseTotal - gcStatsBefore.PauseTotal
	totalAlloc := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
	
	// Print results
	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Total Duration: %v\n", duration)
	fmt.Printf("Operations/sec: %.2f\n", float64(iterations)/duration.Seconds())
	fmt.Println()
	
	fmt.Println("=== Memory Statistics ===")
	fmt.Printf("Total Allocated: %.2f MB\n", float64(totalAlloc)/(1024*1024))
	fmt.Printf("Heap Allocated: %.2f MB\n", float64(memStatsAfter.HeapAlloc)/(1024*1024))
	fmt.Printf("Heap Objects: %d\n", memStatsAfter.HeapObjects)
	fmt.Println()
	
	fmt.Println("=== Garbage Collection Statistics ===")
	fmt.Printf("Number of GCs: %d\n", numGCs)
	fmt.Printf("Total GC Pause: %v\n", totalPause)
	if numGCs > 0 {
		avgPause := totalPause / time.Duration(numGCs)
		fmt.Printf("Average GC Pause: %v\n", avgPause)
		fmt.Printf("GC Pause Overhead: %.2f%%\n", 
			(float64(totalPause)/float64(duration))*100)
	}
	fmt.Printf("Last GC Pause: %v\n", gcStatsAfter.LastPause)
	fmt.Println()
	
	fmt.Println("=== Performance Metrics ===")
	gcCPUFraction := memStatsAfter.GCCPUFraction
	fmt.Printf("GC CPU Fraction: %.2f%%\n", gcCPUFraction*100)
	fmt.Printf("Time per iteration: %v\n", duration/iterations)
	
	// Keep results alive
	_ = results
	
	fmt.Println()
	fmt.Println("Benchmark complete!")
}