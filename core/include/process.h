#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

typedef struct Result {
    bool ret;
    const char *content;
} Result;

typedef struct CommonParams {
    bool endpoint_props;
}CommonParams;

/** JSON body + HTTP status from llama_core HTTP handlers (body malloc'd, or NULL). */
typedef struct LlamaHTTPBody {
    int status;
    char *body;
} LlamaHTTPBody;

bool llama_start(const char * args);
bool llama_stop();
Result llama_gen(int id,const char * js_str);
Result llama_chat(int id,const char * js_str);

bool llama_interactive_start(const char * args,const char * prompt);
bool llama_interactive_stop();

Result whisper_gen(const char * model,const char * input);

bool llama_is_running(void);

CommonParams get_common_params();
LlamaHTTPBody llama_props_http(void);
LlamaHTTPBody llama_slots_http(void);

#ifdef __cplusplus
}
#endif