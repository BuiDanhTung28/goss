// #cgo darwin LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss_c -lfaiss -lstdc++ -lomp -framework Accelerate
package faiss

/*
#include <faiss/c_api/IndexFlat_c.h>
#include <faiss/c_api/Index_c.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"unsafe"
)

// IndexFlat is an index that stores the full vectors and performs exhaustive
// search. This is the most accurate but also the slowest index type.
// It's useful as a baseline for comparison and for small datasets.
type IndexFlat struct {
	Index
}

// NewIndexFlat creates a new flat index with the specified dimension and metric.
// The flat index stores all vectors in memory and performs exhaustive search.
func NewIndexFlat(d int, metric int) (*IndexFlat, error) {
	if d <= 0 {
		return nil, fmt.Errorf("dimension must be positive, got %d", d)
	}

	var cIdx *C.FaissIndex
	if c := C.faiss_IndexFlat_new_with(
		&cIdx,
		C.idx_t(d),
		C.FaissMetricType(metric),
	); c != 0 {
		return nil, wrapError(getLastError(), "IndexFlat creation")
	}

	idx := &faissIndex{idx: cIdx}
	runtime.SetFinalizer(idx, (*faissIndex).Delete)

	return &IndexFlat{idx}, nil
}

// NewIndexFlatIP creates a new flat index with the inner product metric type.
// This is suitable for cosine similarity when vectors are normalized.
func NewIndexFlatIP(d int) (*IndexFlat, error) {
	return NewIndexFlat(d, MetricInnerProduct)
}

// NewIndexFlatL2 creates a new flat index with the L2 (Euclidean) metric type.
// This is the most common metric for similarity search.
func NewIndexFlatL2(d int) (*IndexFlat, error) {
	return NewIndexFlat(d, MetricL2)
}

// NewIndexFlatL1 creates a new flat index with the L1 (Manhattan) metric type.
func NewIndexFlatL1(d int) (*IndexFlat, error) {
	return NewIndexFlat(d, MetricL1)
}

// NewIndexFlatLinf creates a new flat index with the L-infinity metric type.
func NewIndexFlatLinf(d int) (*IndexFlat, error) {
	return NewIndexFlat(d, MetricLinf)
}

// Xb returns the index's vectors.
// The returned slice becomes invalid after any add or remove operation.
// Use with caution as it provides direct access to internal memory.
func (idx *IndexFlat) Xb() []float32 {
	if idx.Index == nil {
		return nil
	}

	var size C.size_t
	var ptr *C.float
	C.faiss_IndexFlat_xb(idx.cPtr(), &ptr, &size)

	if ptr == nil || size == 0 {
		return nil
	}

	return (*[1 << 30]float32)(unsafe.Pointer(ptr))[:size:size]
}

// GetVector returns a copy of the vector at the specified index.
// This is safer than using Xb() as it creates a copy.
func (idx *IndexFlat) GetVector(id int64) ([]float32, error) {
	if idx.Index == nil {
		return nil, errors.New("index is nil")
	}

	if id < 0 || id >= idx.Ntotal() {
		return nil, fmt.Errorf("invalid vector ID: %d (valid range: 0-%d)", id, idx.Ntotal()-1)
	}

	d := idx.D()
	vectors := idx.Xb()
	if vectors == nil {
		return nil, errors.New("no vectors in index")
	}

	start := int(id) * d
	end := start + d

	if end > len(vectors) {
		return nil, errors.New("vector access out of bounds")
	}

	// Create a copy
	result := make([]float32, d)
	copy(result, vectors[start:end])
	return result, nil
}

// GetVectors returns copies of multiple vectors by their IDs.
func (idx *IndexFlat) GetVectors(ids []int64) ([]float32, error) {
	if idx.Index == nil {
		return nil, errors.New("index is nil")
	}

	if len(ids) == 0 {
		return nil, errors.New("empty IDs slice")
	}

	d := idx.D()
	ntotal := idx.Ntotal()

	// Validate all IDs first
	for i, id := range ids {
		if id < 0 || id >= ntotal {
			return nil, fmt.Errorf("invalid vector ID at index %d: %d (valid range: 0-%d)", i, id, ntotal-1)
		}
	}

	vectors := idx.Xb()
	if vectors == nil {
		return nil, errors.New("no vectors in index")
	}

	result := make([]float32, len(ids)*d)
	for i, id := range ids {
		start := int(id) * d
		end := start + d

		if end > len(vectors) {
			return nil, fmt.Errorf("vector access out of bounds for ID %d", id)
		}

		copy(result[i*d:(i+1)*d], vectors[start:end])
	}

	return result, nil
}

// GetVectorRange returns a copy of vectors in the specified range [start, end).
func (idx *IndexFlat) GetVectorRange(start, end int64) ([]float32, error) {
	if idx.Index == nil {
		return nil, errors.New("index is nil")
	}

	if start < 0 || end < 0 {
		return nil, fmt.Errorf("negative range values: start=%d, end=%d", start, end)
	}

	if start >= end {
		return nil, fmt.Errorf("invalid range: start=%d >= end=%d", start, end)
	}

	ntotal := idx.Ntotal()
	if start >= ntotal {
		return nil, fmt.Errorf("start index %d >= ntotal %d", start, ntotal)
	}

	if end > ntotal {
		end = ntotal
	}

	d := idx.D()
	vectors := idx.Xb()
	if vectors == nil {
		return nil, errors.New("no vectors in index")
	}

	count := int(end - start)
	result := make([]float32, count*d)

	for i := 0; i < count; i++ {
		srcStart := int(start+int64(i)) * d
		srcEnd := srcStart + d
		dstStart := i * d
		dstEnd := dstStart + d

		if srcEnd > len(vectors) {
			return nil, fmt.Errorf("vector access out of bounds")
		}

		copy(result[dstStart:dstEnd], vectors[srcStart:srcEnd])
	}

	return result, nil
}

// ComputeDistances computes distances between a query vector and all vectors in the index.
// Returns distances in the same order as the vectors were added.
func (idx *IndexFlat) ComputeDistances(query []float32) ([]float32, error) {
	if idx.Index == nil {
		return nil, errors.New("index is nil")
	}

	d := idx.D()
	if len(query) != d {
		return nil, fmt.Errorf("query dimension %d doesn't match index dimension %d", len(query), d)
	}

	ntotal := idx.Ntotal()
	if ntotal == 0 {
		return nil, errors.New("index is empty")
	}

	// Use search with k = ntotal to get all distances
	distances, _, err := idx.Search(query, ntotal)
	if err != nil {
		return nil, wrapError(err, "compute distances")
	}

	return distances, nil
}

// ComputeDistancesBatch computes distances between multiple query vectors and all vectors in the index
// using SearchBatch for better memory management and performance.
// Returns a matrix where result[i*ntotal+j] is the distance between query i and index vector j.
func (idx *IndexFlat) ComputeDistancesBatch(queries []float32, batchSize int) ([]float32, error) {
	if idx.Index == nil {
		return nil, fmt.Errorf("index is nil")
	}

	d := idx.D()
	if err := ValidateVectors(queries, d); err != nil {
		return nil, wrapError(err, "validate queries")
	}

	ntotal := idx.Ntotal()
	if ntotal == 0 {
		return nil, fmt.Errorf("index is empty")
	}

	if batchSize <= 0 {
		batchSize = DefaultSearchBatchSize
	}

	distances, _, err := idx.SearchBatch(queries, ntotal, batchSize)
	if err != nil {
		return nil, wrapError(err, "compute distances batch")
	}

	numQueries := len(queries) / d
	result := make([]float32, numQueries*int(ntotal))

	for i := 0; i < numQueries; i++ {
		if i < len(distances) && i < len(distances[i]) {
			start := i * int(ntotal)
			end := start + int(ntotal)
			copy(result[start:end], distances[i])
		}
	}

	return result, nil
}

// ComputeL2Norms computes the L2 norms of all vectors in the index.
func (idx *IndexFlat) ComputeL2Norms() ([]float32, error) {
	if idx.Index == nil {
		return nil, fmt.Errorf("index is nil")
	}

	d := idx.D()
	ntotal := idx.Ntotal()

	if ntotal == 0 {
		return nil, fmt.Errorf("index is empty")
	}

	vectors := idx.Xb()
	if vectors == nil {
		return nil, fmt.Errorf("no vectors in index")
	}

	norms := make([]float32, ntotal)

	for i := int64(0); i < ntotal; i++ {
		start := int(i) * d
		end := start + d

		var norm float32
		for j := start; j < end; j++ {
			norm += vectors[j] * vectors[j]
		}

		norms[i] = float32(math.Sqrt(float64(norm)))
	}

	return norms, nil
}

// NormalizeVectors normalizes all vectors in the index to unit length.
// This is useful for converting an L2 index to cosine similarity.
func (idx *IndexFlat) NormalizeVectors() error {
	if idx.Index == nil {
		return errors.New("index is nil")
	}

	norms, err := idx.ComputeL2Norms()
	if err != nil {
		return wrapError(err, "get norms for normalization")
	}

	vectors := idx.Xb()
	if vectors == nil {
		return errors.New("no vectors in index")
	}

	d := idx.D()
	ntotal := idx.Ntotal()

	for i := int64(0); i < ntotal; i++ {
		if norms[i] == 0 {
			continue
		}

		factor := float32(1.0) / norms[i]
		start := int(i) * d
		end := start + d

		// Normalize
		for j := start; j < end; j++ {
			vectors[j] *= factor
		}
	}

	return nil
}

// GetMemoryUsage returns the estimated memory usage of the index in bytes.
func (idx *IndexFlat) GetMemoryUsage() int64 {
	if idx.Index == nil {
		return 0
	}

	d := idx.D()
	ntotal := idx.Ntotal()

	// Each vector is d float32s, each float32 is 4 bytes
	vectorsSize := ntotal * int64(d) * 4

	overhead := int64(1024)

	return vectorsSize + overhead
}

// FlatIndexBuilder helps build flat indices with validation.
type FlatIndexBuilder struct {
	dimension int
	metric    int
	vectors   []float32
	normalize bool
}

// NewFlatIndexBuilder creates a new flat index builder.
func NewFlatIndexBuilder(dimension int) *FlatIndexBuilder {
	return &FlatIndexBuilder{
		dimension: dimension,
		metric:    MetricL2,
		vectors:   make([]float32, 0),
		normalize: false,
	}
}

// SetMetric sets the metric type for the index.
func (b *FlatIndexBuilder) SetMetric(metric int) *FlatIndexBuilder {
	b.metric = metric
	return b
}

// SetNormalize enables/disables vector normalization.
func (b *FlatIndexBuilder) SetNormalize(normalize bool) *FlatIndexBuilder {
	b.normalize = normalize
	return b
}

// AddVector adds a single vector to the builder.
func (b *FlatIndexBuilder) AddVector(vector []float32) *FlatIndexBuilder {
	if len(vector) == b.dimension {
		b.vectors = append(b.vectors, vector...)
	}
	return b
}

// AddVectors adds multiple vectors to the builder.
func (b *FlatIndexBuilder) AddVectors(vectors []float32) *FlatIndexBuilder {
	if len(vectors)%b.dimension == 0 {
		b.vectors = append(b.vectors, vectors...)
	}
	return b
}

// GetVectorCount returns the number of vectors currently in the builder.
func (b *FlatIndexBuilder) GetVectorCount() int {
	return len(b.vectors) / b.dimension
}

// Build creates the flat index with the accumulated vectors.
func (b *FlatIndexBuilder) Build() (*IndexFlat, error) {
	if b.dimension <= 0 {
		return nil, fmt.Errorf("invalid dimension: %d", b.dimension)
	}

	// Create the index
	idx, err := NewIndexFlat(b.dimension, b.metric)
	if err != nil {
		return nil, wrapError(err, "create flat index")
	}

	// Add vectors if any
	if len(b.vectors) > 0 {
		vectors := make([]float32, len(b.vectors))
		copy(vectors, b.vectors)

		// Normalize if requested
		if b.normalize {
			if err := NormalizeVectors(vectors, b.dimension); err != nil {
				idx.Delete()
				return nil, wrapError(err, "normalize vectors")
			}
		}

		// Add vectors to index
		if err := idx.Add(vectors); err != nil {
			idx.Delete()
			return nil, wrapError(err, "add vectors to index")
		}
	}

	return idx, nil
}

// Clear removes all vectors from the builder.
func (b *FlatIndexBuilder) Clear() *FlatIndexBuilder {
	b.vectors = b.vectors[:0]
	return b
}
