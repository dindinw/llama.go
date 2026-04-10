#include <iostream>
#include <thread>
#include <chrono>
#include <cstdlib>
#include <future>

#include "process.h"
#include "./../src/server/server.h"

int main() {
    const char* env_model = "LLAMA_TEST_MODEL";
    const char* model = std::getenv(env_model);

    if (model == nullptr) {
        std::cerr << "error：can't find " << env_model << std::endl;
        return EXIT_FAILURE;
    }

    std::cout << "env: " << env_model << "=" << model << std::endl;

    std::stringstream ss;
    ss << "test_runner -m " << model << " --seed 0";

    std::future<void> start = std::async(std::launch::async, [&](){
        bool ret=llama_start(ss.str().c_str());
        if (!ret) {
            std::cout<<"Start Fail:"<<ret<<std::endl;
        }
    });
    while (!Server::instance().is_running()) {
        int seconds = 5;
        std::cout << "sleep...:"<<seconds<<" seconds" << std::endl;
        std::this_thread::sleep_for(std::chrono::seconds(seconds));
    }

    bool ret =Server::instance().get_health();
    std::cout << "ret:"<<ret<< std::endl;
    ret=llama_stop();

    start.wait();

    if (!ret) {
        return EXIT_FAILURE;
    }
    return EXIT_SUCCESS;
}