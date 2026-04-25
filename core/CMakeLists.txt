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

# Generate embedded web UI assets required by server-http.cpp
set(SERVER_PUBLIC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/llama.cpp/tools/server/public)
set(SERVER_ASSETS_DIR ${CMAKE_BINARY_DIR}/server_assets)
file(MAKE_DIRECTORY ${SERVER_ASSETS_DIR})

foreach(asset index.html.gz loading.html)
    set(input  "${SERVER_PUBLIC_DIR}/${asset}")
    set(output "${SERVER_ASSETS_DIR}/${asset}.hpp")
    list(APPEND GENERATED_ASSETS ${output})
    add_custom_command(
        OUTPUT  "${output}"
        DEPENDS "${input}"
        COMMAND "${CMAKE_COMMAND}"
                "-DINPUT=${input}"
                "-DOUTPUT=${output}"
                -P "${CMAKE_CURRENT_SOURCE_DIR}/llama.cpp/scripts/xxd.cmake"
        COMMENT "Generating ${asset}.hpp"
    )
endforeach()

include_directories(${SERVER_ASSETS_DIR})

# core
set(SRCS
    ${GENERATED_ASSETS}
    src/process_go_bridge_stub.cpp
    src/interactive.cpp
    src/process.cpp
    src/runner.cpp
    src/event_processor.cpp
    src/embedding.cpp
    src/safe_queue.cpp
    src/server.cpp
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
include_directories(./llama.cpp/tools/mtmd)
include_directories(./llama.cpp/tools/server)

link_directories(${CMAKE_BINARY_DIR}/lib)

add_library(${TARGET} STATIC ${SRCS})
target_link_libraries(${TARGET} PRIVATE common llama mtmd)

# test
option(BUILD_TEST "Build the testing tree." ON)

if(BUILD_TEST)
    include(CTest)
    add_subdirectory(tests)
endif()