// #cgo darwin LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss_c -lfaiss -lstdc++ -lomp -framework Accelerate
package faiss

/*
#include <stdlib.h>
#include <faiss/c_api/index_io_c.h>
*/
import "C"
import (
	"errors"
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
		return errors.New("index is nil")
	}

	if fname == "" {
		return errors.New("filename is empty")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fname)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return wrapError(err, "could not create directory")
	}

	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))

	if c := C.faiss_write_index_fname(idx.cPtr(), cfname); c != 0 {
		return wrapError(getLastError(), "write index operation")
	}

	return nil
}

// ReadIndex reads an index from a file.
func ReadIndex(fname string, ioflags int) (Index, error) {
	if fname == "" {
		return nil, errors.New("filename is empty")
	}

	// Check if file exists before calling FAISS API
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return nil, errors.New("index file does not exist")
	}

	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))

	var cIdx *C.FaissIndex
	if c := C.faiss_read_index_fname(cfname, C.int(ioflags), &cIdx); c != 0 {
		return nil, wrapError(getLastError(), "read index operation")
	}
	return NewFaissIndex(cIdx), nil
}
