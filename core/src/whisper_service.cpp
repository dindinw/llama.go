#include "whisper_service.h"

#include <iostream>
#include <string>

WhisperService::WhisperService() {
    std::cout << "Whisper Constructor:"<< std::endl;
}

WhisperService::~WhisperService() {
    std::cout << "Whisper Destructor:"<< std::endl;
}

const std::string WhisperService::generate(const std::string& model,const std::string& input) {

}
