cmake_minimum_required(VERSION 3.14)
project(llama_core)

if(POLICY CMP0077)
    cmake_policy(SET CMP0077 NEW)
endif()

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED true)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED true)

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY_DEBUG ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY_RELEASE ${CMAKE_BINARY_DIR}/lib)
set(EXECUTABLE_OUTPUT_PATH ${CMAKE_BINARY_DIR}/bin)

# llama.cpp
set(GGML_USE_OPENMP OFF)
set(GGML_OPENMP OFF)
set(GGML_STATIC ON)
set(BUILD_SHARED_LIBS OFF)
set(LLAMA_BUILD_COMMON ON)
set(LLAMA_BUILD_TOOLS ON)
set(LLAMA_CURL OFF)
add_subdirectory(llama.cpp ${CMAKE_BINARY_DIR}/llama)

# whisper.cpp — reuse the ggml already built by llama.cpp
set(WHISPER_BUILD_TESTS    OFF)
set(WHISPER_BUILD_EXAMPLES OFF)
set(WHISPER_BUILD_SERVER   OFF)
add_subdirectory(whisper.cpp ${CMAKE_BINARY_DIR}/whisper)

# Generate embedded web UI assets required by server-http.cpp
set(SERVER_PUBLIC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/llama.cpp/tools/server/public)
set(SERVER_ASSETS_DIR ${CMAKE_BINARY_DIR}/server_assets)
file(MAKE_DIRECTORY ${SERVER_ASSETS_DIR})

# loading.html — present in public/ directly
set(_lh_in  "${SERVER_PUBLIC_DIR}/loading.html")
set(_lh_out "${SERVER_ASSETS_DIR}/loading.html.hpp")
add_custom_command(
    OUTPUT  "${_lh_out}"
    DEPENDS "${_lh_in}"
    COMMAND "${CMAKE_COMMAND}" "-DINPUT=${_lh_in}" "-DOUTPUT=${_lh_out}"
            -P "${CMAKE_CURRENT_SOURCE_DIR}/llama.cpp/scripts/xxd.cmake"
    COMMENT "Generating loading.html.hpp"
)

# index.html.gz — compress index.html first, then embed via xxd
set(_html_in "${SERVER_PUBLIC_DIR}/index.html")
set(_gz_file "${SERVER_ASSETS_DIR}/index.html.gz")
set(_gz_out  "${SERVER_ASSETS_DIR}/index.html.gz.hpp")
add_custom_command(
    OUTPUT  "${_gz_out}"
    DEPENDS "${_html_in}"
    COMMAND "${CMAKE_COMMAND}"
            "-DINPUT=${_html_in}"
            "-DOUTPUT=${_gz_file}"
            -P "${CMAKE_CURRENT_SOURCE_DIR}/scripts/gzip_file.cmake"
    COMMAND "${CMAKE_COMMAND}" "-DINPUT=${_gz_file}" "-DOUTPUT=${_gz_out}"
            -P "${CMAKE_CURRENT_SOURCE_DIR}/llama.cpp/scripts/xxd.cmake"
    COMMENT "Compressing index.html and generating index.html.gz.hpp"
)

set(GENERATED_ASSETS "${_lh_out}" "${_gz_out}")
include_directories(${SERVER_ASSETS_DIR})

# core
set(SRCS
    ${GENERATED_ASSETS}
    src/process_go_bridge_stub.cpp
    src/interactive.cpp
    src/process.cpp
    src/event_processor.cpp
    src/embedding.cpp
    src/safe_queue.cpp
    src/server.cpp
    src/whisper_service.cpp
    whisper.cpp/examples/common-whisper.cpp
    whisper.cpp/examples/common.cpp
    llama.cpp/tools/server/server-models.cpp
    llama.cpp/tools/server/server-http.cpp
    llama.cpp/tools/server/server-context.cpp
    llama.cpp/tools/server/server-task.cpp
    llama.cpp/tools/server/server-queue.cpp
    llama.cpp/tools/server/server-common.cpp
)

set(TARGET llama_core)

include_directories(./include)
include_directories(./llama.cpp/include)
include_directories(./llama.cpp/common)
include_directories(./llama.cpp/vendor)
include_directories(./llama.cpp/tools/mtmd)
include_directories(./llama.cpp/tools/server)
include_directories(./whisper.cpp/include)
include_directories(./whisper.cpp/examples)

link_directories(${CMAKE_BINARY_DIR}/lib)

add_library(${TARGET} STATIC ${SRCS})
target_link_libraries(${TARGET} PRIVATE llama-common-base llama-common llama mtmd whisper)

# test
option(BUILD_TEST "Build the testing tree." ON)

if(BUILD_TEST)
    include(CTest)
    add_subdirectory(tests)
endif()