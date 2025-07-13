//go:build windows
// +build windows

package faiss

/*
// CGO flags for Windows with MinGW-w64
#cgo CXXFLAGS: -std=c++17 -O3
#cgo CFLAGS: -I${SRCDIR}/faiss_source
#cgo LDFLAGS: -L${SRCDIR}/internal/lib -lfaiss -lstdc++ -lm
*/
import "C"
