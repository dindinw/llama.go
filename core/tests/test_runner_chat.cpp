#include <chrono>
#include <cstdlib>
#include <future>
#include <iostream>
#include <sstream>
#include <thread>

#include "process.h"
#include "./../src/server/server.h"

namespace {

constexpr int kReadyPollMs = 500;
constexpr int kReadyMaxPolls = 600;

} // namespace

int main() {
    const char * env_model = "LLAMA_TEST_MODEL";
    const char * model = std::getenv(env_model);

    if (model == nullptr) {
        std::cerr << "error: set " << env_model << " to a .gguf path\n";
        return EXIT_FAILURE;
    }

    std::cout << "env: " << env_model << "=" << model << std::endl;

    std::stringstream ss;
    ss << "test_runner_chat -m " << model << " --seed 0";

    std::future<void> start = std::async(std::launch::async, [&]() {
        if (!llama_start(ss.str().c_str())) {
            std::cerr << "llama_start failed\n";
        }
    });

    for (int i = 0; i < kReadyMaxPolls && !Server::instance().is_running(); ++i) {
        std::this_thread::sleep_for(std::chrono::milliseconds(kReadyPollMs));
    }

    if (!Server::instance().is_running()) {
        std::cerr << "error: server did not become ready (timeout)\n";
        (void)llama_stop();
        start.wait();
        return EXIT_FAILURE;
    }

    // OpenAI-style /v1/chat/completions (see oaicompat_chat_params_parse: messages required)
    const char * body = R"({"messages":[{"role":"user","content":"Say OK in one word."}],"max_tokens":8000,"temperature":0,"stream":true})";

    const int req_id = 1;
    Result r = llama_chat(req_id, body);
    std::cout << "llama_chat ret=" << (r.ret ? "true" : "false") << std::endl;

    const bool stopped = llama_stop();
    start.wait();

    if (!r.ret || !stopped) {
        return EXIT_FAILURE;
    }
    return EXIT_SUCCESS;
}
