#pragma once

#include <atomic>
#include <functional>
#include <map>
#include <string>
#include <thread>

// generator-like API for HTTP response generation
// this object response with one of the 2 modes:
// 1) normal response: `data` contains the full response body
// 2) streaming response: each call to next(output) generates the next chunk
//    when next(output) returns false, no more data after the current chunk
//    note: some chunks can be empty, in which case no data is sent for that chunk
struct server_http_res {
    std::string content_type = "application/json; charset=utf-8";
    int status = 200;
    std::string data;
    std::map<std::string, std::string> headers;

    // TODO: move this to a virtual function once we have proper polymorphism support
    std::function<bool(std::string &)> next = nullptr;
    bool is_stream() const {
        return next != nullptr;
    }

    bool is_success() const {
        return status == 200;
    }

    virtual ~server_http_res() = default;
};

// unique pointer, used by set_chunked_content_provider
// httplib requires the stream provider to be stored in heap
using server_http_res_ptr = std::unique_ptr<server_http_res>;

struct server_http_req {
    int id{};
    std::string body{};
    const std::function<bool(int,const std::string&)> write{};
    const std::function<bool()> should_stop=[] { return false; };

    std::map<std::string, std::string> params{}; // path_params + query_params
    std::map<std::string, std::string> headers{}; // reserved for future use
    std::string path{}; // reserved for future use

    std::string get_param(const std::string & key, const std::string & def = "") const {
        auto it = params.find(key);
        if (it != params.end()) {
            return it->second;
        }
        return def;
    }
};

using handler_t = std::function<server_http_res_ptr(const server_http_req & req)>;
