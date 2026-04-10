//
// Created by jinjin on 2026/3/21.
//

#pragma once
#include <vector>
#include <string>
#include "server-context.h"
#include "./../singleton.h"

class Server: public patterns::Singleton<Server> {
    friend class patterns::Singleton<Server>;
private:
    std::unique_ptr<server_routes> routes;
    server_context ctx_server;
    bool running= false;


    Server();
    ~Server();

public:

    bool start(const std::vector<std::string>& args);
    bool stop();

    server_http_res_ptr process(const handler_t& func,const server_http_req& req);
    bool get_health();
    server_http_res_ptr post_completions(const server_http_req& req);
    server_http_res_ptr post_chat_completions(const server_http_req& req);
    server_http_res_ptr get_props(const server_http_req& req);
    server_http_res_ptr get_slots(const server_http_req& req);
    bool is_running() const;
    bool endpoint_props() const;

    /** Drain a completions HTTP response into req.write (e.g. CGO channel); uses req.id. */
    static void flush_http_response_to_sink(const server_http_req & rq, server_http_res & res);
};