# Agent notes

## Build

This repository is Go-first; `llama_core` is built with CMake and linked via cgo.

- Run **`make`** from the **repository root** (same as `./scripts/build.sh`).
- CMake must use **`-S core -B build`**: sources under `core/`, build tree in **`./build`** at the repo root. Do not put the build directory under `core/` (e.g. avoid `core/build_agent`) so artifacts stay aligned with `LD_LIBRARY_PATH` and `build/bin/llama` from the script.
- Cursor CMake Tools: use preset **default** in root `CMakePresets.json` (binary dir `build/`).
