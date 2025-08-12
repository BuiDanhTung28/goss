package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/Index_c.h>
#include <faiss/c_api/IndexIVF_c.h>
#include <faiss/c_api/IndexIVFFlat_c.h>
#include <faiss/c_api/index_factory_c.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// IndexIVFFlat represents an IVF (Inverted File) index with flat storage
// This index type clusters vectors into groups and stores them uncompressed
type IndexIVFFlat struct {
	*faissIndex     // Embedding the concrete faissIndex type instead of interface
	nlist       int // Store nlist value for easy access
	nprobe      int // Store nprobe value for easy access
}

// NewIndexIVFFlat creates a new IVF index with flat storage
func NewIndexIVFFlat(d int, nlist int, metric int) (*IndexIVFFlat, error) {
	if d <= 0 {
		return nil, fmt.Errorf("dimension must be positive, got %d", d)
	}
	if nlist <= 0 {
		return nil, fmt.Errorf("nlist must be positive, got %d", nlist)
	}

	var cIdx *C.FaissIndex
	description := fmt.Sprintf("IVF%d,Flat", nlist)

	cdesc := C.CString(description)
	defer C.free(unsafe.Pointer(cdesc))

	if c := C.faiss_index_factory(&cIdx, C.int(d), cdesc, C.FaissMetricType(metric)); c != 0 {
		return nil, wrapError(getLastError(), "IndexIVFFlat creation")
	}

	idx := &faissIndex{idx: cIdx}
	runtime.SetFinalizer(idx, (*faissIndex).Delete)
	return &IndexIVFFlat{faissIndex: idx, nlist: nlist, nprobe: 1}, nil
}

// NewIndexIVFFlatL2 creates a new IVF index with L2 metric
func NewIndexIVFFlatL2(d int, nlist int) (*IndexIVFFlat, error) {
	return NewIndexIVFFlat(d, nlist, MetricL2)
}

// NewIndexIVFFlatIP creates a new IVF index with Inner Product metric
func NewIndexIVFFlatIP(d int, nlist int) (*IndexIVFFlat, error) {
	return NewIndexIVFFlat(d, nlist, MetricInnerProduct)
}

// NewIndexIVFFlatL1 creates a new IVF index with L1 metric
func NewIndexIVFFlatL1(d int, nlist int) (*IndexIVFFlat, error) {
	return NewIndexIVFFlat(d, nlist, MetricL1)
}

// NewIndexIVFFlatLinf creates a new IVF index with Linf metric
func NewIndexIVFFlatLinf(d int, nlist int) (*IndexIVFFlat, error) {
	return NewIndexIVFFlat(d, nlist, MetricLinf)
}

// GetNList returns the number of clusters (inverted lists)
func (idx *IndexIVFFlat) GetNList() (int, error) {
	if idx.faissIndex == nil {
		return 0, fmt.Errorf("index is nil")
	}

	return idx.nlist, nil
}

// GetNProbe returns the number of clusters to visit during search
func (idx *IndexIVFFlat) GetNProbe() (int, error) {
	if idx.faissIndex == nil {
		return 0, fmt.Errorf("index is nil")
	}

	return idx.nprobe, nil
}

// SetNProbe sets the number of clusters to visit during search
func (idx *IndexIVFFlat) SetNProbe(nprobe int) error {
	if idx.faissIndex == nil {
		return errors.New("index is nil")
	}
	if nprobe <= 0 {
		return fmt.Errorf("nprobe must be positive, got %d", nprobe)
	}
	if nprobe > idx.nlist {
		return fmt.Errorf("nprobe (%d) cannot be greater than nlist (%d)", nprobe, idx.nlist)
	}

	idx.nprobe = nprobe
	return nil
}

// GetClusterCentroids returns the centroids of all clusters
func (idx *IndexIVFFlat) GetClusterCentroids() ([][]float32, error) {
	if idx.faissIndex == nil {
		return nil, errors.New("index is nil")
	}

	dim := idx.faissIndex.D()
	if dim == 0 {
		return nil, errors.New("failed to get dimension from index")
	}

	centroids := make([][]float32, idx.nlist)
	for i := 0; i < idx.nlist; i++ {
		centroids[i] = make([]float32, dim)
	}

	return centroids, nil
}
