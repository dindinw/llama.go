#include "runner.h"
#include <iostream>

#include "arg.h"
#include "common.h"
#include "console.h"
#include "log.h"
#include "sampling.h"
#include "llama.h"
#include "chat.h"
#include "chat.cpp"
#include "message.h"

#include <cstdio>
#include <cstring>
#include <ctime>
#include <fstream>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>

#if defined(_MSC_VER)
#pragma warning(disable: 4244 4267) // possible loss of data
#endif

static void print_usage(int argc, char ** argv) {
    (void) argc;

    LOG("\nexample usage:\n");
    LOG("\n  text generation:     %s -m your_model.gguf -p \"I believe the meaning of life is\" -n 128 -no-cnv\n", argv[0]);
    LOG("\n  chat (conversation): %s -m your_model.gguf -sys \"You are a helpful assistant\"\n", argv[0]);
    LOG("\n");
}

static bool file_exists(const std::string & path) {
    std::ifstream f(path.c_str());
    return f.good();
}

static bool file_is_empty(const std::string & path) {
    std::ifstream f;
    f.exceptions(std::ifstream::failbit | std::ifstream::badbit);
    f.open(path.c_str(), std::ios::in | std::ios::binary | std::ios::ate);
    return f.tellg() == 0;
}

std::string common_chat_formats(
        const struct common_chat_templates * tmpls,
        const std::vector<common_chat_msg> & past_msg,
        const std::vector<common_chat_msg> & new_msg,
        bool use_jinja) {

    common_chat_templates_inputs inputs;
    inputs.use_jinja = use_jinja;
    inputs.add_bos = tmpls->add_bos;
    inputs.add_eos = tmpls->add_eos;

    std::string fmt_past_msg;
    if (!past_msg.empty()) {
        inputs.messages = past_msg;
        inputs.add_generation_prompt = false;
        fmt_past_msg = common_chat_templates_apply(tmpls, inputs).prompt;
    }
    std::ostringstream ss;
    bool add_ass= false;
    if (new_msg[0].content == "user")
        add_ass= true;
    // if the past_msg ends with a newline, we must preserve it in the formatted version
    if (add_ass && !fmt_past_msg.empty() && fmt_past_msg.back() == '\n') {
        ss << "\n";
    };
    // format chat with new_msg
    for (common_chat_msg msg:new_msg) {
        inputs.messages.push_back(msg);
    }

    inputs.add_generation_prompt = add_ass;
    auto fmt_new_msg = common_chat_templates_apply(tmpls, inputs).prompt;
    // get the diff part
    ss << fmt_new_msg.substr(fmt_past_msg.size(), fmt_new_msg.size() - fmt_past_msg.size());
    return ss.str();
}

Runner::Runner() :
    m_params(nullptr),m_model(nullptr),m_smpl(nullptr),m_input_tokens(nullptr),m_output_ss(nullptr),m_output_tokens(nullptr) {
}

Runner::~Runner() {
    std::cout << "Runner Destructor:"<<m_id<< std::endl;
}

bool Runner::start(int id,const std::vector<std::string>& args,bool async,const std::string& prt) {
    return true;
}

bool Runner::stop() {
    if (!isRunning()) {
        std::cout << "No Start:"<<m_id<< std::endl;
        return false;
    }
    std::cout << "Runner Stop:"<<m_id<< std::endl;

    m_running = false;
    m_eprocessor.stop();

    return true;
}

const std::string Runner::generate(const std::string& prompt) {
    if (!isRunning()) {
        std::cout << "No Start:"<<m_id<< std::endl;
        return "";
    }
    std::cout << "Runner generate id:"<<m_id<<" prompt:"<<prompt<< std::endl;

    std::vector<Message> mgs;
    Message mg{"user",prompt};
    mgs.push_back(mg);

    return m_eprocessor.enqueue(mgs);
}

const std::string Runner::chat(const std::vector<Message>& mgs) {
    if (!isRunning()) {
        std::cout << "No Start:"<<m_id<< std::endl;
        return "";
    }
    std::cout << "Runner chat id:"<<m_id<<" message.size:"<<mgs.size()<< std::endl;

    return m_eprocessor.enqueue(mgs);
}

int Runner::getID() {
    return m_id;
}

bool Runner::isRunning() {
    return m_running;
}

bool Runner::getPrompt(EventProcessor::Event& event) {
    if (!isRunning()) {
        return false;
    }
    if (m_async) {
        if (!m_output_ss->str().empty()) {
            try {
                event.result.set_value(m_output_ss->str());
            } catch (...) {
                event.result.set_exception(std::current_exception());
            }
            m_output_ss->str("");
            m_output_ss->clear();
        }
        return m_eprocessor.dequeue(event);
    }
    std::string line;
    std::string buffer;
    bool another_line = true;
    do {
        another_line = console::readline(line, m_params->multiline_input);
        buffer += line;
    } while (another_line);

    std::vector<Message> mgs;
    Message mg{"user",buffer};
    mgs.push_back(mg);

    event.data=mgs;
    return true;
}