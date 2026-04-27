#include "build-info.h"

extern "C" {
    const char* get_llama_build_info() {
        return llama_build_info();
    }
}
