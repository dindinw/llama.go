//go:build darwin
// +build darwin

package wrapper

/*
#cgo CFLAGS: -std=c11
#cgo CXXFLAGS: -std=c++17
#cgo CFLAGS: -I${SRCDIR}/../core/include
#cgo CXXFLAGS: -I${SRCDIR}/../core/include
#cgo LDFLAGS: -framework Foundation -framework Metal -framework MetalKit -framework Accelerate -lstdc++
#cgo LDFLAGS: -L${SRCDIR}/../build/lib -lllama_core -lllama -lllama-common-base -lllama-common -lcpp-httplib -lssl -lcrypto -lmtmd -lggml -lggml-base -lggml-cpu -lggml-blas -lggml-metal
*/
import "C"
