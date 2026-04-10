#include "process.h"
#include "log.h"
#include "whisper_service.h"
#include "server/server.h"

#include <cstdlib>
#include <cstring>
#include <sstream>
#include <string>

extern "C" {
    void PushToChan(int id, const char* val);
    void CloseChan(int id);
}

bool llama_start(const char * args) {
    if (Server::instance().is_running()) {
        return false;
    }

    std::istringstream iss(args);
    std::vector<std::string> v_args;
    std::string v_a;
    while (iss >> v_a) {
        v_args.push_back(v_a);
    }

    if (!Server::instance().start(v_args)) {
        return false;
    }
    return true;
}

bool llama_stop() {
    if (!Server::instance().is_running()) {
        return false;
    }
    if (!Server::instance().stop()) {
        return false;
    }
    return true;
}

Result llama_gen(int id, const char * js_str) {
    if (!Server::instance().is_running()) {
        return {false, nullptr};
    }
    if (!js_str) {
        return {false, nullptr};
    }

    server_http_req rq{
            id,
            std::string(js_str),
            [](int cid, const std::string & content) {
                PushToChan(cid, content.c_str());
                return true;
            }
    };

    server_http_res_ptr rp = Server::instance().post_completions(rq);
    const bool ok = rp->is_success();

    CloseChan(id);
    return {ok, nullptr};
}

Result llama_chat(int id, const char * js_str) {
    if (!Server::instance().is_running()) {
        return {false, nullptr};
    }
    if (!js_str) {
        return {false, nullptr};
    }

    server_http_req rq{
            id,
            std::string(js_str),
            [](int cid, const std::string & content) {
                PushToChan(cid, content.c_str());
                return true;
            }
    };

    server_http_res_ptr rp = Server::instance().post_chat_completions(rq);
    const bool ok = rp->is_success();

    CloseChan(id);
    return {ok, nullptr};
}

Result whisper_gen(const char * model,const char * input) {
    (void)model;
    (void)input;
    return {true, nullptr};
}

static LlamaHTTPBody make_http_body(const server_http_res_ptr &rp) {
    LlamaHTTPBody out{};
    if (!rp) {
        out.status = 500;
        return out;
    }
    out.status = rp->status;
    if (!rp->data.empty()) {
        out.body = strdup(rp->data.c_str());
    }
    return out;
}

extern "C" {

bool llama_is_running(void) {
    return Server::instance().is_running();
}

CommonParams get_common_params(void) {
    if (!Server::instance().is_running()) {
        return {false};
    }
    return {Server::instance().endpoint_props()};
}

LlamaHTTPBody llama_props_http(void) {
    LlamaHTTPBody out{};
    out.status = 503;
    if (!Server::instance().is_running()) {
        return out;
    }
    server_http_req req{};
    return make_http_body(Server::instance().get_props(req));
}

LlamaHTTPBody llama_slots_http(void) {
    LlamaHTTPBody out{};
    out.status = 503;
    if (!Server::instance().is_running()) {
        return out;
    }
    server_http_req req{};
    return make_http_body(Server::instance().get_slots(req));
}

}