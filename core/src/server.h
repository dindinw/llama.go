//
// Created by jinjin on 2026/3/21.
//

#pragma once
#include <vector>
#include <string>
#include "singleton.h"
#include "server-http.h"    // upstream, resolved via include_directories
#include "server-context.h" // upstream, resolved via include_directories

// Convenience alias: our legacy code used a global handler_t.
using handler_t = server_http_context::handler_t;

// Returns a reference to a static "never stop" predicate.
// The upstream server_http_req holds should_stop as a const-ref member,
// so callers that don't have a natural stop signal can use this.
inline const std::function<bool()> & server_never_stop() {
    static const std::function<bool()> fn = [] { return false; };
    return fn;
}

// Constructs a server_http_req with only a body; all other fields are defaults.
inline server_http_req make_server_req(std::string body = {}) {
    // field order: params, headers, path, query_string, body, files, should_stop
    return server_http_req{{}, {}, {}, {}, std::move(body), {}, server_never_stop()};
}

// CGO write sink: carries the channel id and the streaming push callback.
// Kept separate from server_http_req so the upstream struct stays unmodified.
struct server_write_sink {
    int id = 0;
    std::function<bool(int, const std::string &)> write;
    explicit operator bool() const { return static_cast<bool>(write); }
};

class Server: public patterns::Singleton<Server> {
    friend class patterns::Singleton<Server>;
private:
    std::unique_ptr<server_routes> routes;
    std::unique_ptr<server_context> ctx_server;
    bool running = false;
    bool _endpoint_props = false;

    Server();
    ~Server();

    static void flush_to_sink(const server_write_sink & sink, server_http_res & res);

public:

    bool start(const std::vector<std::string>& args);
    bool stop();

    server_http_res_ptr process(const handler_t & func, const server_http_req & req, const server_write_sink & sink = {});
    bool get_health();
    server_http_res_ptr post_completions(const server_http_req & req, const server_write_sink & sink = {});
    server_http_res_ptr post_chat_completions(const server_http_req & req, const server_write_sink & sink = {});
    server_http_res_ptr get_props(const server_http_req & req);
    server_http_res_ptr get_slots(const server_http_req & req);
    bool is_running() const;
    bool endpoint_props() const;
};
