//go:build linux
// +build linux

package faiss

/*
#cgo CXXFLAGS: -std=c++17 -O3
#cgo CFLAGS: -I${SRCDIR}/faiss_source
#cgo LDFLAGS: -L${SRCDIR}/internal/lib -lfaiss -lstdc++ -lm -lrt
// On Linux, OpenMP is usually found with -fopenmp
#cgo LDFLAGS: -fopenmp
*/
import "C"
