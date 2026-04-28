#!/usr/bin/env bash

set -e

if [ ! -d "./core" ]; then
    echo "Must run in the root directory of the project"
    exit 1
fi

if [ ! -d "./core/llama.cpp/src" ]; then
    git submodule update --init --recursive
    echo "Update llama.cpp"
fi

cmake --version


coreDir=$(pwd)/core
buildDir=$(pwd)/build

echo "core dir:" ${coreDir}
echo "build dir:" ${buildDir}

# open-ssl
if [[ "$(uname -s)" == "Linux" ]]; then
   if ! pkg-config --exists openssl 2>/dev/null; then
     sudo apt-get update
     sudo apt-get install -y pkg-config libssl-dev
   fi
fi

# cuda
cudaTag=""
cudaCmake=""
if [[ "$(uname -s)" == "Linux" ]]; then
    if [[ -d "/usr/local/cuda" ]] && command -v nvcc &> /dev/null; then
        echo "Try use CUDA"
        cudaCmake="-DGGML_CUDA=ON"
    fi
fi

cmake -DCMAKE_BUILD_TYPE=Release $cudaCmake -G "Unix Makefiles" -S $coreDir -B $buildDir
cmake --build $buildDir --target llama_core -- -j 9

if [ -e $buildDir/lib/libggml-cuda.a ]; then
    cudaTag="-tags=cuda"
fi

# go
go version

GITVER=$(git rev-parse --short=7 HEAD)
GITDIRTY=$(git diff --quiet || echo '-dirty')
GITVERSION="${GITVER}${GITDIRTY}"
versionBuild="github.com/Qitmeer/llama.go/version.Build=dev-${GITVERSION}"

export CGO_ENABLED=1
export LD_LIBRARY_PATH=$buildDir/lib

# On macOS, cpp-httplib (built with OpenSSL) needs OpenSSL link/include flags.
# Prefer pkg-config, fallback to Homebrew prefix.
if [[ "$(uname -s)" == "Darwin" ]]; then
    openssl_cflags=""
    openssl_ldflags=""

    if command -v pkg-config &> /dev/null; then
        if pkg-config --exists openssl; then
            # Only inject search paths; actual -lssl/-lcrypto come from cgo files.
            openssl_cflags="$(pkg-config --cflags-only-I openssl)"
            openssl_ldflags="$(pkg-config --libs-only-L openssl)"
        elif pkg-config --exists openssl@3; then
            openssl_cflags="$(pkg-config --cflags-only-I openssl@3)"
            openssl_ldflags="$(pkg-config --libs-only-L openssl@3)"
        fi
    fi

    if [[ -z "${openssl_ldflags}" ]] && command -v brew &> /dev/null; then
        brew_prefix="$(brew --prefix openssl@3 2>/dev/null || true)"
        if [[ -z "${brew_prefix}" ]]; then
            brew_prefix="$(brew --prefix openssl 2>/dev/null || true)"
        fi
        if [[ -n "${brew_prefix}" ]]; then
            openssl_cflags="-I${brew_prefix}/include"
            openssl_ldflags="-L${brew_prefix}/lib"
        fi
    fi

    if [[ -n "${openssl_cflags}" ]]; then
        export CGO_CFLAGS="${CGO_CFLAGS:-} ${openssl_cflags}"
    fi
    if [[ -n "${openssl_ldflags}" ]]; then
        export CGO_LDFLAGS="${CGO_LDFLAGS:-} ${openssl_ldflags}"
    fi
fi

cd ./cmd/llama
go build $cudaTag -ldflags "-X ${versionBuild}" -o $buildDir/bin/llama

echo "Output executable file:${buildDir}/bin/llama"
$buildDir/bin/llama version


