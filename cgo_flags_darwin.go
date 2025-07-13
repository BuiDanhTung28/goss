//go:build darwin
// +build darwin

package faiss

/*
#cgo CXXFLAGS: -std=c++17 -O3
#cgo CFLAGS: -I${SRCDIR}/faiss_source
#cgo LDFLAGS: -L${SRCDIR}/internal/lib -lfaiss -lstdc++ -lm
// Link with libomp installed via homebrew
#cgo LDFLAGS: -L/opt/homebrew/opt/libomp/lib -lomp
*/
import "C"
