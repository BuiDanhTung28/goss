//go:build darwin
// +build darwin

package faiss

/*
#cgo CXXFLAGS: -std=c++17 -O3
#cgo CFLAGS: -I${SRCDIR}/faiss_source
#cgo darwin LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss_c -lfaiss -lstdc++ -lomp
// Link with libomp installed via homebrew
#cgo LDFLAGS: -L/opt/homebrew/opt/libomp/lib -lomp
*/
import "C"
