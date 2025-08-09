//go:build darwin
// +build darwin

package faiss

/*
#cgo CXXFLAGS: -std=c++17 -O3
#cgo CFLAGS: -I${SRCDIR}/faiss_source
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/internal/lib/darwin_arm64 -lfaiss -lstdc++
// Link with libomp installed via homebrew
#cgo LDFLAGS: -L/opt/homebrew/opt/libomp/lib -lomp
*/
import "C"
