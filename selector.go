package faiss

/*
#include <faiss/c_api/impl/AuxIndexStructures_c.h>
*/
import "C"
import (
	"fmt"
	"runtime"
	"sort"
)

// IDSelector represents a set of IDs to remove from an index.
// It provides different strategies for selecting which vectors to remove.
type IDSelector struct {
	sel *C.FaissIDSelector
}

// NewIDSelectorRange creates a selector that removes IDs in the range [imin, imax).
// This is useful for removing a contiguous range of vector IDs.
func NewIDSelectorRange(imin, imax int64) (*IDSelector, error) {
	if imin < 0 || imax < 0 {
		return nil, fmt.Errorf("invalid range: imin=%d, imax=%d (must be non-negative)", imin, imax)
	}
	if imin >= imax {
		return nil, fmt.Errorf("invalid range: imin=%d >= imax=%d", imin, imax)
	}

	var sel *C.FaissIDSelectorRange
	c := C.faiss_IDSelectorRange_new(&sel, C.idx_t(imin), C.idx_t(imax))
	if c != 0 {
		return nil, wrapError(getLastError(), "IDSelectorRange creation")
	}

	selector := &IDSelector{(*C.FaissIDSelector)(sel)}
	runtime.SetFinalizer(selector, (*IDSelector).Delete)
	return selector, nil
}

// NewIDSelectorBatch creates a selector that removes specific IDs from a batch.
// The indices slice contains the specific vector IDs to remove.
func NewIDSelectorBatch(indices []int64) (*IDSelector, error) {
	if len(indices) == 0 {
		return nil, fmt.Errorf("empty indices slice")
	}

	// Validate indices
	for i, id := range indices {
		if id < 0 {
			return nil, fmt.Errorf("invalid ID at index %d: %d (must be non-negative)", i, id)
		}
	}

	var sel *C.FaissIDSelectorBatch
	if c := C.faiss_IDSelectorBatch_new(
		&sel,
		C.size_t(len(indices)),
		(*C.idx_t)(&indices[0]),
	); c != 0 {
		return nil, wrapError(getLastError(), "IDSelectorBatch creation")
	}

	selector := &IDSelector{(*C.FaissIDSelector)(sel)}
	runtime.SetFinalizer(selector, (*IDSelector).Delete)
	return selector, nil
}

// NewIDSelectorAnd creates a selector that removes IDs that match ALL of the provided selectors.
// This is useful for complex filtering where multiple conditions must be met.
func NewIDSelectorAnd(selectors ...*IDSelector) (*IDSelector, error) {
	if len(selectors) == 0 {
		return nil, fmt.Errorf("at least one selector required")
	}

	for i, sel := range selectors {
		if sel == nil || sel.sel == nil {
			return nil, fmt.Errorf("selector at index %d is nil", i)
		}
	}

	// For simplicity, we'll implement this as a batch selector
	// In a real implementation, this would require additional C bindings
	return nil, fmt.Errorf("IDSelectorAnd not implemented - requires additional C bindings")
}

// NewIDSelectorOr creates a selector that removes IDs that match ANY of the provided selectors.
// This is useful for complex filtering where any condition can trigger removal.
func NewIDSelectorOr(selectors ...*IDSelector) (*IDSelector, error) {
	if len(selectors) == 0 {
		return nil, fmt.Errorf("at least one selector required")
	}

	for i, sel := range selectors {
		if sel == nil || sel.sel == nil {
			return nil, fmt.Errorf("selector at index %d is nil", i)
		}
	}

	// For simplicity, we'll implement this as a batch selector
	// In a real implementation, this would require additional C bindings
	return nil, fmt.Errorf("IDSelectorOr not implemented - requires additional C bindings")
}

// NewIDSelectorNot creates a selector that removes IDs that do NOT match the provided selector.
// This is useful for inverse selection.
func NewIDSelectorNot(selector *IDSelector, ntotal int64) (*IDSelector, error) {
	if selector == nil || selector.sel == nil {
		return nil, fmt.Errorf("selector is nil")
	}

	if ntotal <= 0 {
		return nil, fmt.Errorf("ntotal must be positive")
	}

	// For simplicity, we'll implement this as a batch selector
	// In a real implementation, this would require additional C bindings
	return nil, fmt.Errorf("IDSelectorNot not implemented - requires additional C bindings")
}

// Delete frees the memory associated with the selector.
func (s *IDSelector) Delete() {
	if s.sel != nil {
		C.faiss_IDSelector_free(s.sel)
		s.sel = nil
	}
	runtime.SetFinalizer(s, nil)
}

// IsNil checks if the selector is nil or has been deleted
func (s *IDSelector) IsNil() bool {
	return s == nil || s.sel == nil
}

// Utility functions for working with ID selectors

// ValidateIDs validates a slice of IDs
func ValidateIDs(ids []int64, maxID int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("empty IDs slice")
	}

	for i, id := range ids {
		if id < 0 {
			return fmt.Errorf("negative ID at index %d: %d", i, id)
		}
		if maxID >= 0 && id >= maxID {
			return fmt.Errorf("ID at index %d (%d) >= maxID (%d)", i, id, maxID)
		}
	}

	return nil
}

// RemoveDuplicateIDs removes duplicate IDs from a slice and returns a sorted slice
func RemoveDuplicateIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return ids
	}

	// Sort first
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	// Remove duplicates
	result := make([]int64, 0, len(ids))
	result = append(result, ids[0])

	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[i-1] {
			result = append(result, ids[i])
		}
	}

	return result
}

// CreateBatchSelector creates a batch selector with duplicate removal and validation
func CreateBatchSelector(ids []int64, maxID int64) (*IDSelector, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("empty IDs slice")
	}

	cleanIDs := RemoveDuplicateIDs(ids)

	// Validate
	if err := ValidateIDs(cleanIDs, maxID); err != nil {
		return nil, wrapError(err, "ID validation")
	}

	return NewIDSelectorBatch(cleanIDs)
}

// CreateRangeSelector creates a range selector with validation
func CreateRangeSelector(start, end int64, maxID int64) (*IDSelector, error) {
	if start < 0 || end < 0 {
		return nil, fmt.Errorf("negative range values: start=%d, end=%d", start, end)
	}

	if start >= end {
		return nil, fmt.Errorf("invalid range: start=%d >= end=%d", start, end)
	}

	if maxID >= 0 && start >= maxID {
		return nil, fmt.Errorf("range start (%d) >= maxID (%d)", start, maxID)
	}

	// Clamp end to maxID if necessary
	if maxID >= 0 && end > maxID {
		end = maxID
	}

	return NewIDSelectorRange(start, end)
}

// BatchSelectorBuilder helps build complex batch selectors
type BatchSelectorBuilder struct {
	ids    []int64
	maxID  int64
	sorted bool
}

// NewBatchSelectorBuilder creates a new batch selector builder
func NewBatchSelectorBuilder() *BatchSelectorBuilder {
	return &BatchSelectorBuilder{
		ids:    make([]int64, 0),
		maxID:  -1,
		sorted: false,
	}
}

// SetMaxID sets the maximum allowed ID
func (b *BatchSelectorBuilder) SetMaxID(maxID int64) *BatchSelectorBuilder {
	b.maxID = maxID
	return b
}

// AddID adds a single ID to the selector
func (b *BatchSelectorBuilder) AddID(id int64) *BatchSelectorBuilder {
	b.ids = append(b.ids, id)
	b.sorted = false
	return b
}

// AddIDs adds multiple IDs to the selector
func (b *BatchSelectorBuilder) AddIDs(ids ...int64) *BatchSelectorBuilder {
	b.ids = append(b.ids, ids...)
	b.sorted = false
	return b
}

// AddRange adds a range of IDs to the selector
func (b *BatchSelectorBuilder) AddRange(start, end int64) *BatchSelectorBuilder {
	for i := start; i < end; i++ {
		b.ids = append(b.ids, i)
	}
	b.sorted = false
	return b
}

// Build creates the ID selector
func (b *BatchSelectorBuilder) Build() (*IDSelector, error) {
	if len(b.ids) == 0 {
		return nil, fmt.Errorf("no IDs added to selector")
	}

	return CreateBatchSelector(b.ids, b.maxID)
}

// Count returns the number of IDs currently in the builder
func (b *BatchSelectorBuilder) Count() int {
	return len(b.ids)
}

// Clear removes all IDs from the builder
func (b *BatchSelectorBuilder) Clear() *BatchSelectorBuilder {
	b.ids = b.ids[:0]
	b.sorted = false
	return b
}

// GetIDs returns a copy of the current IDs
func (b *BatchSelectorBuilder) GetIDs() []int64 {
	result := make([]int64, len(b.ids))
	copy(result, b.ids)
	return result
}
