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
//
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

	// RangeSearch queries the index with the vectors in x.
	// Returns all vectors with distance < radius.
	RangeSearch(x []float32, radius float32) (*RangeSearchResult, error)

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

func (idx *faissIndex) RangeSearch(x []float32, radius float32) (
	*RangeSearchResult, error,
) {
	if idx.idx == nil {
		return nil, ErrNullPointer
	}

	d := idx.D()
	if err := ValidateVectors(x, d); err != nil {
		return nil, wrapError(err, "range_search vectors validation")
	}

	if err := ValidateRadius(radius); err != nil {
		return nil, wrapError(err, "range_search radius validation")
	}

	if !idx.IsTrained() {
		return nil, wrapError(ErrIndexNotTrained, "range_search operation")
	}

	n := len(x) / d
	var rsr *C.FaissRangeSearchResult
	if c := C.faiss_RangeSearchResult_new(&rsr, C.idx_t(n)); c != 0 {
		return nil, wrapError(getLastError(), "range_search result creation")
	}

	if c := C.faiss_Index_range_search(
		idx.idx,
		C.idx_t(n),
		(*C.float)(&x[0]),
		C.float(radius),
		rsr,
	); c != 0 {
		C.faiss_RangeSearchResult_free(rsr)
		return nil, wrapError(getLastError(), "range_search operation")
	}

	return NewRangeSearchResult(rsr), nil
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

// RangeSearchResult is the result of a range search.
type RangeSearchResult struct {
	rsr *C.FaissRangeSearchResult
}

// NewRangeSearchResult creates a new range search result wrapper
func NewRangeSearchResult(rsr *C.FaissRangeSearchResult) *RangeSearchResult {
	result := &RangeSearchResult{rsr: rsr}
	runtime.SetFinalizer(result, (*RangeSearchResult).Delete)
	return result
}

// Nq returns the number of queries.
func (r *RangeSearchResult) Nq() int {
	if r.rsr == nil {
		return 0
	}
	return int(C.faiss_RangeSearchResult_nq(r.rsr))
}

// Lims returns a slice containing start and end indices for queries in the
// distances and labels slices returned by Labels.
func (r *RangeSearchResult) Lims() []int {
	if r.rsr == nil {
		return nil
	}

	var lims *C.size_t
	C.faiss_RangeSearchResult_lims(r.rsr, &lims)
	length := r.Nq() + 1
	return (*[1 << 30]int)(unsafe.Pointer(lims))[:length:length]
}

// Labels returns the unsorted IDs and respective distances for each query.
// The result for query i is labels[lims[i]:lims[i+1]].
func (r *RangeSearchResult) Labels() (labels []int64, distances []float32) {
	if r.rsr == nil {
		return nil, nil
	}

	lims := r.Lims()
	if len(lims) == 0 {
		return nil, nil
	}

	length := lims[len(lims)-1]
	var clabels *C.idx_t
	var cdist *C.float
	C.faiss_RangeSearchResult_labels(r.rsr, &clabels, &cdist)

	if clabels != nil {
		labels = (*[1 << 30]int64)(unsafe.Pointer(clabels))[:length:length]
	}
	if cdist != nil {
		distances = (*[1 << 30]float32)(unsafe.Pointer(cdist))[:length:length]
	}
	return
}

// Delete frees the memory associated with r.
func (r *RangeSearchResult) Delete() {
	if r.rsr != nil {
		C.faiss_RangeSearchResult_free(r.rsr)
		r.rsr = nil
	}
	runtime.SetFinalizer(r, nil)
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
