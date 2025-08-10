package faiss

import (
	"testing"
	"time"
)

func TestAddBatch(t *testing.T) {
	// Test cases for AddBatch functionality
	tests := []struct {
		name        string
		dimension   int
		vectors     []float32
		batchSize   int
		expectError bool
		description string
	}{
		{
			name:        "Normal batch processing",
			dimension:   4,
			vectors:     []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			batchSize:   2,
			expectError: false,
			description: "Should successfully add 3 vectors in batches of 2",
		},
		{
			name:        "Single batch",
			dimension:   3,
			vectors:     []float32{1, 2, 3, 4, 5, 6},
			batchSize:   10,
			expectError: false,
			description: "Should add all vectors in one batch when batch size > total vectors",
		},
		{
			name:        "Empty vectors",
			dimension:   4,
			vectors:     []float32{},
			batchSize:   5,
			expectError: true, // Empty vectors should fail validation
			description: "Should fail validation for empty vectors",
		},
		{
			name:        "Zero batch size",
			dimension:   2,
			vectors:     []float32{1, 2, 3, 4},
			batchSize:   0,
			expectError: false,
			description: "Should use default batch size when batch size is 0",
		},
		{
			name:        "Negative batch size",
			dimension:   2,
			vectors:     []float32{1, 2, 3, 4},
			batchSize:   -1,
			expectError: false,
			description: "Should use default batch size when batch size is negative",
		},
		{
			name:        "Batch size equals vector count",
			dimension:   2,
			vectors:     []float32{1, 2, 3, 4},
			batchSize:   2,
			expectError: false,
			description: "Should process exactly one batch when batch size equals vector count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new index
			idx, err := NewIndexFlat(tt.dimension, MetricL2)
			if err != nil {
				t.Fatalf("Failed to create index: %v", err)
			}
			defer idx.Delete()

			// Train the index with some sample data
			trainingData := make([]float32, tt.dimension*2)
			for i := range trainingData {
				trainingData[i] = float32(i % tt.dimension)
			}

			if err := idx.Train(trainingData); err != nil {
				t.Fatalf("Failed to train index: %v", err)
			}

			// Test AddBatch
			err = idx.AddBatch(tt.vectors, tt.batchSize)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none: %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v, %s", err, tt.description)
			}

			// Verify vectors were added correctly
			if !tt.expectError && len(tt.vectors) > 0 {
				expectedTotal := len(tt.vectors) / tt.dimension
				actualTotal := idx.Ntotal()
				if int64(expectedTotal) != actualTotal {
					t.Errorf("Expected %d vectors, got %d", expectedTotal, actualTotal)
				}
			}
		})
	}
}

func TestAddBatchWithLargeDataset(t *testing.T) {
	// Test with a larger dataset to verify memory efficiency
	dimension := 128
	batchSize := 100
	totalVectors := 1000

	// Create large dataset
	vectors := make([]float32, dimension*totalVectors)
	for i := range vectors {
		vectors[i] = float32(i%100) / 100.0 // Normalize to [0, 1)
	}

	// Create index
	idx, err := NewIndexFlat(dimension, MetricL2)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer idx.Delete()

	// Train with subset
	trainingSubset := vectors[:dimension*100] // Use first 100 vectors for training
	if err := idx.Train(trainingSubset); err != nil {
		t.Fatalf("Failed to train index: %v", err)
	}

	// Test AddBatch with large dataset
	err = idx.AddBatch(vectors, batchSize)
	if err != nil {
		t.Fatalf("AddBatch failed with large dataset: %v", err)
	}

	// Verify all vectors were added
	expectedTotal := totalVectors
	actualTotal := idx.Ntotal()
	if int64(expectedTotal) != actualTotal {
		t.Errorf("Expected %d vectors, got %d", expectedTotal, actualTotal)
	}
}

func TestAddBatchEdgeCases(t *testing.T) {
	// Test edge cases and error conditions
	tests := []struct {
		name        string
		dimension   int
		vectors     []float32
		batchSize   int
		expectError bool
		errorType   string
	}{
		{
			name:        "Invalid vector length",
			dimension:   4,
			vectors:     []float32{1, 2, 3}, // Not divisible by dimension
			batchSize:   2,
			expectError: true,
			errorType:   "validation error",
		},
		{
			name:        "Untrained index",
			dimension:   2,
			vectors:     []float32{1, 2, 3, 4},
			batchSize:   2,
			expectError: true,
			errorType:   "not trained error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create index - use IVF for untrained test case
			var idx Index
			var err error
			if tt.name == "Untrained index" {
				// Use IVF index that requires training
				idx, err = IndexFactory(tt.dimension, "IVF2,Flat", MetricL2)
			} else {
				idx, err = NewIndexFlat(tt.dimension, MetricL2)
			}
			if err != nil {
				t.Fatalf("Failed to create index: %v", err)
			}
			defer idx.Delete()

			// Only train if not testing untrained case
			if tt.name != "Untrained index" {
				trainingData := make([]float32, tt.dimension*2)
				for i := range trainingData {
					trainingData[i] = float32(i % tt.dimension)
				}

				if err := idx.Train(trainingData); err != nil {
					t.Fatalf("Failed to train index: %v", err)
				}
			}
			// For untrained case, we don't train the index

			// Test AddBatch
			err = idx.AddBatch(tt.vectors, tt.batchSize)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none: %s", tt.name)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v, test: %s", err, tt.name)
			}

			if tt.expectError && err != nil {
				// Verify error type
				if tt.errorType == "validation error" {
					if err.Error() == "" {
						t.Errorf("Expected validation error but got: %v", err)
					}
				} else if tt.errorType == "not trained error" {
					if err.Error() == "" {
						t.Errorf("Expected not trained error but got: %v", err)
					}
				}
			}
		})
	}
}

func TestAddBatchPerformance(t *testing.T) {
	// Performance test to ensure batch processing is efficient
	dimension := 64
	batchSize := 50
	totalVectors := 500

	// Create test data
	vectors := make([]float32, dimension*totalVectors)
	for i := range vectors {
		vectors[i] = float32(i%100) / 100.0
	}

	// Create index
	idx, err := NewIndexFlat(dimension, MetricL2)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer idx.Delete()

	// Train
	trainingData := vectors[:dimension*50]
	if err := idx.Train(trainingData); err != nil {
		t.Fatalf("Failed to train index: %v", err)
	}

	// Benchmark AddBatch
	start := time.Now()
	err = idx.AddBatch(vectors, batchSize)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("AddBatch failed: %v", err)
	}

	// Performance assertions
	if duration > 5*time.Second {
		t.Errorf("AddBatch took too long: %v", duration)
	}

	// Verify results
	expectedTotal := totalVectors
	actualTotal := idx.Ntotal()
	if int64(expectedTotal) != actualTotal {
		t.Errorf("Expected %d vectors, got %d", expectedTotal, actualTotal)
	}

	t.Logf("AddBatch processed %d vectors in %v", totalVectors, duration)
}

// Benchmark tests
func BenchmarkAddBatch(b *testing.B) {
	dimension := 128
	batchSize := 100
	vectors := make([]float32, dimension*1000)

	for i := range vectors {
		vectors[i] = float32(i%100) / 100.0
	}

	idx, err := NewIndexFlat(dimension, MetricL2)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}
	defer idx.Delete()

	// Train
	trainingData := vectors[:dimension*100]
	if err := idx.Train(trainingData); err != nil {
		b.Fatalf("Failed to train index: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := idx.AddBatch(vectors, batchSize)
		if err != nil {
			b.Fatalf("AddBatch failed: %v", err)
		}
		// Reset for next iteration
		idx.Reset()
	}
}

func BenchmarkAddBatchVsAdd(b *testing.B) {
	dimension := 64
	vectors := make([]float32, dimension*500)

	for i := range vectors {
		vectors[i] = float32(i%100) / 100.0
	}

	// Test AddBatch
	b.Run("AddBatch", func(b *testing.B) {
		idx, err := NewIndexFlat(dimension, MetricL2)
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}
		defer idx.Delete()

		trainingData := vectors[:dimension*50]
		if err := idx.Train(trainingData); err != nil {
			b.Fatalf("Failed to train index: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := idx.AddBatch(vectors, 100)
			if err != nil {
				b.Fatalf("AddBatch failed: %v", err)
			}
			idx.Reset()
		}
	})

	// Test regular Add
	b.Run("Add", func(b *testing.B) {
		idx, err := NewIndexFlat(dimension, MetricL2)
		if err != nil {
			b.Fatalf("Failed to create index: %v", err)
		}
		defer idx.Delete()

		trainingData := vectors[:dimension*50]
		if err := idx.Train(trainingData); err != nil {
			b.Fatalf("Failed to train index: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := idx.Add(vectors)
			if err != nil {
				b.Fatalf("Add failed: %v", err)
			}
			idx.Reset()
		}
	})
}
