package faiss

import (
	"fmt"
	"os"
	"sync"
)

// PersistentIndex is a wrapper around an Index that automatically persists
// changes to a file. It is safe for concurrent use.
type PersistentIndex struct {
	Index
	path string
	mu   sync.RWMutex
}

// NewPersistentIndex creates or loads a persistent index from a file.
// If the file does not exist, a new index is created using the provided factory function.
func NewPersistentIndex(path string, factory func() (Index, error)) (*PersistentIndex, error) {
	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File does not exist, create a new index
		idx, err := factory()
		if err != nil {
			return nil, fmt.Errorf("factory error: %w", err)
		}
		return &PersistentIndex{Index: idx, path: path}, nil
	} else if err != nil {
		// Another error occurred
		return nil, fmt.Errorf("stat error: %w", err)
	}

	// File exists, read it
	idx, err := ReadIndex(path, 0)
	if err != nil {
		return nil, fmt.Errorf("read index error: %w", err)
	}
	return &PersistentIndex{Index: idx, path: path}, nil
}

// Add adds vectors to the index and persists the changes to the file.
func (p *PersistentIndex) Add(x []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Index.Add(x); err != nil {
		return err
	}

	// Persist the entire index to the file.
	return WriteIndex(p.Index, p.path)
}

// AddWithIDs adds vectors with their own IDs and persists the changes.
func (p *PersistentIndex) AddWithIDs(x []float32, xids []int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.Index.AddWithIDs(x, xids); err != nil {
		return err
	}

	return WriteIndex(p.Index, p.path)
}

// RemoveIDs removes vectors and persists the changes.
func (p *PersistentIndex) RemoveIDs(sel *IDSelector) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n, err := p.Index.RemoveIDs(sel)
	if err != nil {
		return 0, err
	}

	if err := WriteIndex(p.Index, p.path); err != nil {
		// This part is tricky. The removal from memory was successful,
		// but saving to disk failed. We return the number of removed items
		// but also the disk error.
		return n, err
	}

	return n, nil
}
