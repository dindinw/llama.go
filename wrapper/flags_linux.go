//go:build linux && !cuda

package wrapper

/*
#cgo CFLAGS: -std=c11
#cgo CXXFLAGS: -std=c++17
#cgo CFLAGS: -I${SRCDIR}/../core/include
#cgo CXXFLAGS: -I${SRCDIR}/../core/include
#cgo LDFLAGS: -L${SRCDIR}/../build/lib -lllama_core -lllama-common-base -lllama-common -lcpp-httplib -lssl -lcrypto -lllama -lmtmd -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm
*/
import "C"
