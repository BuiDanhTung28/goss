// #cgo darwin LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss_c -lfaiss -lstdc++ -lomp -framework Accelerate
// #cgo linux LDFLAGS: -L${SRCDIR}/internal/lib/linux_x64 -lfaiss_c -lfaiss -lstdc++ -lomp
//
//go:generate ./build.sh
package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/Index_c.h>
#include <faiss/c_api/index_io_c.h>
#include <faiss/c_api/impl/AuxIndexStructures_c.h>
#include <faiss/c_api/index_factory_c.h>
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// Index is a Faiss index for vector similarity search.
// Note that some index implementations do not support all methods.
// Check the Faiss wiki to see what operations an index supports.
// https://github.com/facebookresearch/faiss/wiki
type Index interface {
	// D returns the dimension of the indexed vectors.
	D() int
	// IsTrained returns true if the index has been trained or does not require
	// training.
	IsTrained() bool

	// Ntotal returns the number of indexed vectors.
	Ntotal() int64

	// MetricType returns the metric type of the index.
	MetricType() int

	// Train trains the index on a representative set of vectors.
	// Some index types require training before vectors can be added.
	Train(x []float32) error

	// Add adds vectors to the index.
	// The vectors are stored with sequential IDs starting from the current Ntotal.
	Add(x []float32) error

	// AddWithIDs is like Add, but stores xids instead of sequential IDs.
	// This allows custom ID assignment for vectors.
	AddWithIDs(x []float32, xids []int64) error

	// Search queries the index with the vectors in x.
	// Returns the IDs of the k nearest neighbors for each query vector and the
	// corresponding distances.
	Search(x []float32, k int64) (distances []float32, labels []int64, err error)

	// SearchBatch queries the index with multiple vectors in batches
	// Returns distances and labels for each query vector
	SearchBatch(queries []float32, k int64, batchSize int) (distances [][]float32, labels [][]int64, err error)

	// AddBatch adds vectors in batches for better memory management and performance
	AddBatch(vectors []float32, batchSize int) error

	// Reset removes all vectors from the index.
	Reset() error

	// RemoveIDs removes the vectors specified by sel from the index.
	// Returns the number of elements removed and error.
	RemoveIDs(sel *IDSelector) (int, error)

	// Delete frees the memory used by the index.
	Delete()

	// Internal method to get C pointer
	cPtr() *C.FaissIndex
}

// faissIndex is the main implementation of the Index interface
type faissIndex struct {
	idx *C.FaissIndex
}

// NewFaissIndex creates a new index wrapper around a C FaissIndex
func NewFaissIndex(cIdx *C.FaissIndex) Index {
	idx := &faissIndex{idx: cIdx}
	runtime.SetFinalizer(idx, (*faissIndex).Delete)
	return idx
}

func (idx *faissIndex) cPtr() *C.FaissIndex {
	return idx.idx
}

func (idx *faissIndex) D() int {
	if idx.idx == nil {
		return 0
	}
	return int(C.faiss_Index_d(idx.idx))
}

func (idx *faissIndex) IsTrained() bool {
	if idx.idx == nil {
		return false
	}
	return C.faiss_Index_is_trained(idx.idx) != 0
}

func (idx *faissIndex) Ntotal() int64 {
	if idx.idx == nil {
		return 0
	}
	return int64(C.faiss_Index_ntotal(idx.idx))
}

func (idx *faissIndex) MetricType() int {
	if idx.idx == nil {
		return MetricL2
	}
	return int(C.faiss_Index_metric_type(idx.idx))
}

func (idx *faissIndex) Train(x []float32) error {
	if idx.idx == nil {
		return ErrNullPointer
	}

	d := idx.D()
	if err := ValidateVectors(x, d); err != nil {
		return wrapError(err, "train vectors validation")
	}

	n := len(x) / d
	if c := C.faiss_Index_train(idx.idx, C.idx_t(n), (*C.float)(&x[0])); c != 0 {
		return wrapError(getLastError(), "train operation")
	}
	return nil
}

func (idx *faissIndex) Add(x []float32) error {
	if idx.idx == nil {
		return ErrNullPointer
	}

	d := idx.D()
	if err := ValidateVectors(x, d); err != nil {
		return wrapError(err, "add vectors validation")
	}

	if !idx.IsTrained() {
		return wrapError(ErrIndexNotTrained, "add operation")
	}

	n := len(x) / d
	if c := C.faiss_Index_add(idx.idx, C.idx_t(n), (*C.float)(&x[0])); c != 0 {
		return wrapError(getLastError(), "add operation")
	}
	return nil
}

func (idx *faissIndex) AddWithIDs(x []float32, xids []int64) error {
	if idx.idx == nil {
		return ErrNullPointer
	}

	d := idx.D()
	if err := ValidateVectors(x, d); err != nil {
		return wrapError(err, "add_with_ids vectors validation")
	}

	if !idx.IsTrained() {
		return wrapError(ErrIndexNotTrained, "add_with_ids operation")
	}

	n := len(x) / d
	if len(xids) != n {
		return wrapError(fmt.Errorf("number of IDs (%d) doesn't match number of vectors (%d)", len(xids), n), "add_with_ids")
	}

	if c := C.faiss_Index_add_with_ids(
		idx.idx,
		C.idx_t(n),
		(*C.float)(&x[0]),
		(*C.idx_t)(&xids[0]),
	); c != 0 {
		return wrapError(getLastError(), "add_with_ids operation")
	}
	return nil
}

func (idx *faissIndex) Search(x []float32, k int64) (
	distances []float32, labels []int64, err error,
) {
	if idx.idx == nil {
		return nil, nil, ErrNullPointer
	}

	d := idx.D()
	if err := ValidateVectors(x, d); err != nil {
		return nil, nil, wrapError(err, "search vectors validation")
	}

	if err := ValidateK(k); err != nil {
		return nil, nil, wrapError(err, "search k validation")
	}

	if !idx.IsTrained() {
		return nil, nil, wrapError(ErrIndexNotTrained, "search operation")
	}

	n := len(x) / d
	distances = make([]float32, int64(n)*k)
	labels = make([]int64, int64(n)*k)

	if c := C.faiss_Index_search(
		idx.idx,
		C.idx_t(n),
		(*C.float)(&x[0]),
		C.idx_t(k),
		(*C.float)(&distances[0]),
		(*C.idx_t)(&labels[0]),
	); c != 0 {
		err = wrapError(getLastError(), "search operation")
		return nil, nil, err
	}
	return
}

func (idx *faissIndex) SearchBatch(queries []float32, k int64, batchSize int) (distances [][]float32, labels [][]int64, err error) {
	if idx.idx == nil {
		return nil, nil, ErrNullPointer
	}

	if batchSize <= 0 {
		batchSize = DefaultSearchBatchSize
	}

	d := idx.D()
	if err := ValidateVectors(queries, d); err != nil {
		return nil, nil, wrapError(err, "search batch queries validation")
	}

	if !idx.IsTrained() {
		return nil, nil, wrapError(ErrIndexNotTrained, "search batch operation")
	}

	totalQueries := len(queries) / d
	if totalQueries == 0 {
		return make([][]float32, 0), make([][]int64, 0), nil
	}

	// Use optimal batch size if the provided one is too large
	if batchSize > totalQueries {
		batchSize = totalQueries
	}

	// Initialize result slices
	distances = make([][]float32, totalQueries)
	labels = make([][]int64, totalQueries)

	// Process in batches
	for i := 0; i < totalQueries; i += batchSize {
		end := i + batchSize
		if end > totalQueries {
			end = totalQueries
		}

		batchStart := i * d
		batchEnd := end * d
		batch := queries[batchStart:batchEnd]

		// Search this batch using existing Search method
		batchDistances, batchLabels, err := idx.Search(batch, k)
		if err != nil {
			return nil, nil, wrapError(err, fmt.Sprintf("search batch %d-%d", i, end-1))
		}

		// Distribute results to final slices
		for j := 0; j < end-i; j++ {
			queryIdx := i + j
			start := j * int(k)
			end := start + int(k)

			distances[queryIdx] = batchDistances[start:end]
			labels[queryIdx] = batchLabels[start:end]
		}
	}

	return distances, labels, nil
}

func (idx *faissIndex) AddBatch(vectors []float32, batchSize int) error {
	if idx.idx == nil {
		return ErrNullPointer
	}

	if batchSize <= 0 {
		batchSize = DefaultAddBatchSize
	}

	d := idx.D()
	if err := ValidateVectors(vectors, d); err != nil {
		return wrapError(err, "add batch vectors validation")
	}

	if !idx.IsTrained() {
		return wrapError(ErrIndexNotTrained, "add batch operation")
	}

	totalVectors := len(vectors) / d
	if totalVectors == 0 {
		return nil // Nothing to add
	}

	// Use optimal batch size if the provided one is too large
	if batchSize > totalVectors {
		batchSize = totalVectors
	}

	for i := 0; i < totalVectors; i += batchSize {
		end := i + batchSize
		if end > totalVectors {
			end = totalVectors
		}

		batchStart := i * d
		batchEnd := end * d
		batch := vectors[batchStart:batchEnd]

		if err := idx.Add(batch); err != nil {
			return wrapError(err, fmt.Sprintf("add batch %d-%d", i, end-1))
		}
	}

	return nil
}

func (idx *faissIndex) Reset() error {
	if idx.idx == nil {
		return ErrNullPointer
	}

	if c := C.faiss_Index_reset(idx.idx); c != 0 {
		return wrapError(getLastError(), "reset operation")
	}
	return nil
}

func (idx *faissIndex) RemoveIDs(sel *IDSelector) (int, error) {
	if idx.idx == nil {
		return 0, ErrNullPointer
	}

	if sel == nil || sel.sel == nil {
		return 0, wrapError(ErrNullPointer, "remove_ids selector")
	}

	var nRemoved C.size_t
	if c := C.faiss_Index_remove_ids(idx.idx, sel.sel, &nRemoved); c != 0 {
		return 0, wrapError(getLastError(), "remove_ids operation")
	}
	return int(nRemoved), nil
}

func (idx *faissIndex) Delete() {
	if idx.idx != nil {
		C.faiss_Index_free(idx.idx)
		idx.idx = nil
	}
	runtime.SetFinalizer(idx, nil)
}

// IndexFactory builds a composite index using the factory pattern.
// description is a comma-separated list of components.
// Common descriptions:
//   - "Flat" - Exact search
//   - "IVF100,Flat" - IVF with 100 centroids
//   - "IVF100,PQ8" - IVF with 100 centroids and 8-bit PQ
//   - "HNSW32" - HNSW with 32 connections per node
func IndexFactory(d int, description string, metric int) (Index, error) {
	if d <= 0 {
		return nil, ErrInvalidDimension
	}

	if description == "" {
		description = "Flat"
	}

	cdesc := C.CString(description)
	defer C.free(unsafe.Pointer(cdesc))

	var cIdx *C.FaissIndex
	c := C.faiss_index_factory(&cIdx, C.int(d), cdesc, C.FaissMetricType(metric))
	if c != 0 {
		return nil, wrapError(getLastError(), "index factory")
	}

	idx := &faissIndex{idx: cIdx}
	runtime.SetFinalizer(idx, (*faissIndex).Delete)

	return idx, nil
}
