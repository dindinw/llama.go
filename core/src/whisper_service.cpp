#include "whisper_service.h"

#include "whisper.h"
#include "common-whisper.h"

#include <iostream>
#include <string>
#include <thread>
#include <vector>

WhisperService::WhisperService() {}
WhisperService::~WhisperService() {}

const std::string WhisperService::generate(const std::string& model, const std::string& input) {
    struct whisper_context_params cparams = whisper_context_default_params();
    cparams.use_gpu    = true;
    cparams.flash_attn = true;

    struct whisper_context * ctx = whisper_init_from_file_with_params(model.c_str(), cparams);
    if (!ctx) {
        std::cerr << "WhisperService: failed to init context from model: " << model << std::endl;
        return "";
    }

    std::vector<float> pcmf32;
    std::vector<std::vector<float>> pcmf32s;
    if (!read_audio_data(input, pcmf32, pcmf32s, false)) {
        std::cerr << "WhisperService: failed to read audio file: " << input << std::endl;
        whisper_free(ctx);
        return "";
    }

    whisper_full_params wparams = whisper_full_default_params(WHISPER_SAMPLING_GREEDY);
    wparams.print_realtime   = false;
    wparams.print_progress   = false;
    wparams.print_timestamps = false;
    wparams.print_special    = false;
    wparams.translate        = false;
    wparams.language         = "auto";
    wparams.n_threads        = std::min(4, (int) std::thread::hardware_concurrency());
    wparams.no_timestamps    = true;

    if (whisper_full(ctx, wparams, pcmf32.data(), (int) pcmf32.size()) != 0) {
        std::cerr << "WhisperService: failed to process audio" << std::endl;
        whisper_free(ctx);
        return "";
    }

    const int n_segments = whisper_full_n_segments(ctx);
    std::string result;
    for (int i = 0; i < n_segments; i++) {
        result += whisper_full_get_segment_text(ctx, i);
    }

    whisper_free(ctx);
    return result;
}
