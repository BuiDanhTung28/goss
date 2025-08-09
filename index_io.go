// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss -lstdc++
package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/index_io_c.h>
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

// IO flags for index reading/writing
const (
	IOFlagMmap     = C.FAISS_IO_FLAG_MMAP      // Memory-map the index file
	IOFlagReadOnly = C.FAISS_IO_FLAG_READ_ONLY // Open in read-only mode
)

// WriteIndex writes an index to a file.
func WriteIndex(idx Index, fname string) error {
	if idx == nil {
		return fmt.Errorf("index is nil")
	}

	if fname == "" {
		return fmt.Errorf("filename is empty")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fname)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create directory: %w", err)
	}

	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))

	if c := C.faiss_write_index_fname(idx.cPtr(), cfname); c != 0 {
		return getLastError()
	}

	return nil
}

// ReadIndex reads an index from a file.
func ReadIndex(fname string, ioflags int) (Index, error) {
	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))

	var cIdx *C.FaissIndex
	if c := C.faiss_read_index_fname(cfname, C.int(ioflags), &cIdx); c != 0 {
		return nil, getLastError()
	}
	return NewFaissIndex(cIdx), nil
}
