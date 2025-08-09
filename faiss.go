// #cgo darwin LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss_c -lfaiss -lstdc++ -lomp
//
//go:generate ./build.sh
package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/Index_c.h>
#include <faiss/c_api/index_io_c.h>
#include <faiss/c_api/error_c.h>
*/
import "C"
import (
	"errors"
	"fmt"
)

// Error handling
var (
	ErrInvalidDimension = errors.New("invalid dimension")
	ErrInvalidK         = errors.New("invalid k value")
	ErrInvalidRadius    = errors.New("invalid radius")
	ErrEmptyVectors     = errors.New("empty vectors")
	ErrIndexNotTrained  = errors.New("index not trained")
	ErrNullPointer      = errors.New("null pointer")
)

// getLastError returns the last error from the Faiss C API
func getLastError() error {
	errMsg := C.faiss_get_last_error()
	if errMsg == nil {
		return errors.New("unknown FAISS error")
	}
	return errors.New(C.GoString(errMsg))
}

// wrapError wraps a FAISS error with additional context
func wrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Metric types for similarity computation
const (
	MetricInnerProduct  = C.METRIC_INNER_PRODUCT // Inner product (cosine for normalized vectors)
	MetricL2            = C.METRIC_L2            // L2 (Euclidean) distance
	MetricL1            = C.METRIC_L1            // L1 (Manhattan) distance
	MetricLinf          = C.METRIC_Linf          // L-infinity distance
	MetricLp            = C.METRIC_Lp            // Lp distance
	MetricCanberra      = C.METRIC_Canberra      // Canberra distance
	MetricBrayCurtis    = C.METRIC_BrayCurtis    // Bray-Curtis distance
	MetricJensenShannon = C.METRIC_JensenShannon // Jensen-Shannon divergence
)

// Index types for factory creation
const (
	IndexTypeFlat    = "Flat"
	IndexTypeIVF     = "IVF"
	IndexTypeIVFFlat = "IVFFlat"
	IndexTypeIVFPQ   = "IVFPQ"
	IndexTypeHNSW    = "HNSW"
	IndexTypeLSH     = "LSH"
	IndexTypePQ      = "PQ"
)

// Common index configurations
const (
	DefaultNList        = 100 // Default number of clusters for IVF
	DefaultNProbe       = 1   // Default number of probes for search
	DefaultM            = 8   // Default number of sub-vectors for PQ
	DefaultNBits        = 8   // Default bits per sub-vector for PQ
	DefaultHNSWM        = 16  // Default number of connections for HNSW
	DefaultHNSWEfSearch = 16  // Default search parameter for HNSW
)

// Utility functions

// ValidateVectors validates that vectors have the correct dimensions
func ValidateVectors(vectors []float32, d int) error {
	if len(vectors) == 0 {
		return ErrEmptyVectors
	}
	if d <= 0 {
		return ErrInvalidDimension
	}
	if len(vectors)%d != 0 {
		return fmt.Errorf("vectors length %d is not divisible by dimension %d", len(vectors), d)
	}
	return nil
}

// ValidateK validates the k parameter for search
func ValidateK(k int64) error {
	if k <= 0 {
		return ErrInvalidK
	}
	return nil
}

// ValidateRadius validates the radius parameter for range search
func ValidateRadius(radius float32) error {
	if radius < 0 {
		return ErrInvalidRadius
	}
	return nil
}

// NormalizeVectors normalizes vectors to unit length (for cosine similarity)
func NormalizeVectors(vectors []float32, d int) error {
	if err := ValidateVectors(vectors, d); err != nil {
		return err
	}

	n := len(vectors) / d
	for i := 0; i < n; i++ {
		start := i * d
		end := start + d

		// Calculate norm
		norm := float32(0)
		for j := start; j < end; j++ {
			norm += vectors[j] * vectors[j]
		}

		if norm == 0 {
			continue // Skip zero vectors
		}

		norm = float32(1.0) / float32(sqrt(float64(norm)))

		// Normalize
		for j := start; j < end; j++ {
			vectors[j] *= norm
		}
	}

	return nil
}

// sqrt computes square root
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// GetVectorBatch extracts a batch of vectors from a larger slice
func GetVectorBatch(vectors []float32, d int, start, count int) []float32 {
	if start < 0 || count <= 0 {
		return nil
	}

	n := len(vectors) / d
	if start >= n {
		return nil
	}

	if start+count > n {
		count = n - start
	}

	startIdx := start * d
	endIdx := startIdx + count*d

	return vectors[startIdx:endIdx]
}

// CreateIndexDescription creates a description string for IndexFactory
func CreateIndexDescription(indexType string, params map[string]interface{}) string {
	switch indexType {
	case IndexTypeFlat:
		return "Flat"
	case IndexTypeIVFFlat:
		nlist := DefaultNList
		if v, ok := params["nlist"]; ok {
			if n, ok := v.(int); ok {
				nlist = n
			}
		}
		return fmt.Sprintf("IVF%d,Flat", nlist)
	case IndexTypeIVFPQ:
		nlist := DefaultNList
		m := DefaultM
		nbits := DefaultNBits
		if v, ok := params["nlist"]; ok {
			if n, ok := v.(int); ok {
				nlist = n
			}
		}
		if v, ok := params["m"]; ok {
			if n, ok := v.(int); ok {
				m = n
			}
		}
		if v, ok := params["nbits"]; ok {
			if n, ok := v.(int); ok {
				nbits = n
			}
		}
		return fmt.Sprintf("IVF%d,PQ%dx%d", nlist, m, nbits)
	case IndexTypeHNSW:
		M := DefaultHNSWM
		if v, ok := params["M"]; ok {
			if n, ok := v.(int); ok {
				M = n
			}
		}
		return fmt.Sprintf("HNSW%d", M)
	default:
		return indexType
	}
}

// GetDefaultMetricType returns the default metric type for a given index type
func GetDefaultMetricType(indexType string) int {
	switch indexType {
	case IndexTypeFlat, IndexTypeIVFFlat, IndexTypeIVFPQ:
		return MetricL2
	case IndexTypeHNSW:
		return MetricL2
	default:
		return MetricL2
	}
}

// EstimateMemoryUsage estimates memory usage for an index
func EstimateMemoryUsage(indexType string, d int, n int64, params map[string]interface{}) int64 {
	switch indexType {
	case IndexTypeFlat:
		return n * int64(d) * 4 // 4 bytes per float32
	case IndexTypeIVFFlat:
		return n * int64(d) * 4 // Similar to flat for vectors
	case IndexTypeIVFPQ:
		m := DefaultM
		if v, ok := params["m"]; ok {
			if mv, ok := v.(int); ok {
				m = mv
			}
		}
		return n * int64(m) // Compressed representation
	case IndexTypeHNSW:
		M := DefaultHNSWM
		if v, ok := params["M"]; ok {
			if mv, ok := v.(int); ok {
				M = mv
			}
		}
		return n * (int64(d)*4 + int64(M)*8) // Vectors + connections
	default:
		return n * int64(d) * 4 // Default estimation
	}
}
