OS := $(shell uname)
RM := rm -rf
BUILD_DIR := ./build

all: run

.PHONY: run clean

run:
ifeq ($(OS), Linux)
	@echo "Running on Linux"
	@./scripts/build.sh
else ifeq ($(OS), Darwin)
	@echo "Running on macOS"
	@./scripts/build.sh
else
	@echo "Running on Windows or Unknown OS"
	@powershell -ExecutionPolicy Bypass -File ./scripts/build_windows.ps1
endif

test:
ifeq ($(OS), Linux)
	@echo "Running test on Linux"
	@echo "TODO ./scripts/test.sh"
else ifeq ($(OS), Darwin)
	@echo "Running test on macOS"
	@./scripts/test.sh
else
	@echo "Running test on Windows or Unknown OS"
	@echo "TODO ./scripts/test.sh"
endif

clean-bin:
	$(RM) $(BUILD_DIR)/bin
clean:
	@echo "Cleaning $(BUILD_DIR) directory"
ifeq ($(OS), Linux)
	$(RM) $(BUILD_DIR)
else ifeq ($(OS), Darwin)
	$(RM) $(BUILD_DIR)
else
	@powershell -Command "if (Test-Path '$(BUILD_DIR)') { Remove-Item -Recurse -Force '$(BUILD_DIR)' }"
endif
