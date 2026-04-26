# gzip_file.cmake — compress a file using the system gzip tool
# Usage: cmake -DINPUT=<src> -DOUTPUT=<dst.gz> -P gzip_file.cmake
#
# Using execute_process with OUTPUT_FILE avoids shell-redirect quoting issues
# that arise when cmake generates Makefile rules for add_custom_command.

if(NOT INPUT OR NOT OUTPUT)
    message(FATAL_ERROR "gzip_file.cmake: INPUT and OUTPUT must be set")
endif()

execute_process(
    COMMAND gzip -9 -c "${INPUT}"
    OUTPUT_FILE "${OUTPUT}"
    RESULT_VARIABLE _result
)

if(_result)
    message(FATAL_ERROR "gzip failed (exit ${_result}) for: ${INPUT}")
endif()
