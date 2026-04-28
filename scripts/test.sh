#!/usr/bin/env bash

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

    go test -v ./...
fi

if [[ "$(uname -s)" == "Linux" ]]; then
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

    if [[ -n "${openssl_cflags}" ]]; then
        export CGO_CFLAGS="${CGO_CFLAGS:-} ${openssl_cflags}"
    fi
    if [[ -n "${openssl_ldflags}" ]]; then
        export CGO_LDFLAGS="${CGO_LDFLAGS:-} ${openssl_ldflags}"
    fi

    if [[ -d "/usr/local/cuda" ]] && command -v nvcc &> /dev/null; then
        tags="-tags cuda"
    fi

    go test $tags -v ./...
fi


