#include "server.h"

#include "arg.h"
#include "common.h"
#include "llama.h"
#include "log.h"

#include <exception>
#include <thread> // for std::thread::hardware_concurrency
#include <string>

void Server::flush_to_sink(const server_write_sink & sink, server_http_res & res) {
    if (!sink) {
        return;
    }
    const int id = sink.id;
    if (res.is_stream()) {
        std::string chunk;
        while (res.next) {
            const bool more = res.next(chunk);
            if (!chunk.empty() && !sink.write(id, chunk)) {
                break;
            }
            chunk.clear();
            if (!more) {
                break;
            }
        }
    } else if (!res.data.empty()) {
        sink.write(id, res.data);
    }
}

// wrapper function that handles exceptions and logs errors
// this is to make sure handler_t never throws exceptions; instead, it returns an error response
static handler_t ex_wrapper(handler_t func) {
    return [func = std::move(func)](const server_http_req & req) -> server_http_res_ptr {
        std::string message;
        error_type error;
        try {
            return func(req);
        } catch (const std::invalid_argument & e) {
            error = ERROR_TYPE_INVALID_REQUEST;
            message = e.what();
        } catch (const std::exception & e) {
            error = ERROR_TYPE_SERVER;
            message = e.what();
        } catch (...) {
            error = ERROR_TYPE_SERVER;
            message = "unknown error";
        }

        auto res = std::make_unique<server_http_res>();
        res->status = 500;
        try {
            json error_data = format_error_response(message, error);
            res->status = json_value(error_data, "code", 500);
            res->data = safe_json_to_str({{ "error", error_data }});
            SRV_WRN("got exception: %s\n", res->data.c_str());
        } catch (const std::exception & e) {
            SRV_ERR("got another exception: %s | while handling exception: %s\n", e.what(), message.c_str());
            res->data = "Internal Server Error";
        }
        return res;
    };
}

Server::Server() {
    LOG_INF("%s: Server()\n", __func__);
}

Server::~Server() {
    LOG_INF("%s: ~Server()\n", __func__);
}

bool Server::start(const std::vector<std::string>& args) {

    LOG_INF("%s: Server().start, %d\n", __func__, int(args.size()));
    std::vector<char*> v_argv;
    for (auto& t : args) {
        v_argv.push_back(const_cast<char*>(t.c_str()));
    }
    int argc = int(v_argv.size());

    common_params params;

    if (!common_params_parse(argc, v_argv.data(), params, LLAMA_EXAMPLE_SERVER)) {
        return false;
    }

    if (params.embedding && params.n_batch > params.n_ubatch) {
        LOG_WRN("%s: embeddings enabled with n_batch (%d) > n_ubatch (%d)\n", __func__, params.n_batch, params.n_ubatch);
        LOG_WRN("%s: setting n_batch = n_ubatch = %d to avoid assertion failure\n", __func__, params.n_ubatch);
        params.n_batch = params.n_ubatch;
    }

    if (params.n_parallel < 0) {
        LOG_INF("%s: n_parallel is set to auto, using n_parallel = 4 and kv_unified = true\n", __func__);
        params.n_parallel = 4;
        params.kv_unified = true;
    }

    if (params.model_alias.empty() && !params.model.name.empty()) {
        params.model_alias = {params.model.name};
    }

    _endpoint_props = params.endpoint_props;

    common_init();

    llama_backend_init();
    llama_numa_init(params.numa);

    LOG_INF("system info: n_threads = %d, n_threads_batch = %d, total_threads = %d\n",
            params.cpuparams.n_threads, params.cpuparams_batch.n_threads,
            std::thread::hardware_concurrency());
    LOG_INF("\n");
    LOG_INF("%s\n", common_params_get_system_info(params).c_str());
    LOG_INF("\n");

    ctx_server = std::make_unique<server_context>();
    routes     = std::make_unique<server_routes>(params, *ctx_server);

    // load the model
    LOG_INF("%s: loading model\n", __func__);

    if (!ctx_server->load_model(params)) {
        LOG_ERR("%s: exiting due to model loading error\n", __func__);
        routes.reset();
        ctx_server.reset();
        llama_backend_free();
        return false;
    }

    routes->update_meta(*ctx_server);

    LOG_INF("%s: model loaded\n", __func__);
    LOG_INF("%s: starting the main loop...\n", __func__);
    running = true;

    // blocks until ctx_server->terminate() is called
    ctx_server->start_loop();

    auto * ll_ctx = ctx_server->get_llama_context();
    if (ll_ctx != nullptr) {
        llama_memory_breakdown_print(ll_ctx);
    }

    SRV_INF("%s: cleaning up before exit...\n", __func__);

    // Destroy routes first: it holds a const-ref to server_context_impl.
    routes.reset();

    // Destroy server_context next: its destructor calls server_context_impl::destroy(),
    // which frees all GGML/Metal buffers (KV cache, model buffers, compute buffers).
    // This MUST happen before llama_backend_free() on macOS 15+ to satisfy
    // ggml_metal_rsets_free's assertion that all residency sets are empty.
    ctx_server.reset();

    llama_backend_free();

    return true;
}

bool Server::stop() {
    if (ctx_server) {
        ctx_server->terminate();
    }
    running = false;
    return true;
}

bool Server::is_running() const {
    return running;
}

server_http_res_ptr Server::process(const handler_t & func, const server_http_req & req, const server_write_sink & sink) {
    auto handler = ex_wrapper(func);
    server_http_res_ptr response = handler(req);
    flush_to_sink(sink, *response);
    LOG_INF("%s: Server().process, %s\n", __func__, response->data.c_str());
    return response;
}

bool Server::get_health() {
    server_http_req shr = make_server_req();
    server_http_res_ptr response = process(routes->get_health, shr);
    return response->status == 200;
}

server_http_res_ptr Server::post_completions(const server_http_req & req, const server_write_sink & sink) {
    return process(routes->post_completions_oai, req, sink);
}

server_http_res_ptr Server::post_chat_completions(const server_http_req & req, const server_write_sink & sink) {
    return process(routes->post_chat_completions, req, sink);
}

server_http_res_ptr Server::get_props(const server_http_req & req) {
    return process(routes->get_props, req);
}

server_http_res_ptr Server::get_slots(const server_http_req & req) {
    return process(routes->get_slots, req);
}

bool Server::endpoint_props() const {
    return _endpoint_props;
}
