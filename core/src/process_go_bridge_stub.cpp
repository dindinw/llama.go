// Go/cgo exports PushToChan/CloseChan from wrapper/bridge.go. Pure C++ links
// (e.g. core tests) need definitions; weak stubs are overridden when linking
// with the cgo object that provides the real implementations.

#if defined(__GNUC__) || defined(__clang__)
#define LLAMA_BRIDGE_STUB_WEAK __attribute__((weak))
#else
#define LLAMA_BRIDGE_STUB_WEAK
#endif

extern "C" {

LLAMA_BRIDGE_STUB_WEAK void PushToChan(int id, const char * val) {
    (void)id;
    (void)val;
}

LLAMA_BRIDGE_STUB_WEAK void CloseChan(int id) {
    (void)id;
}

}
