GO = go

# Detect CPU architecture
ifeq ($(shell uname -m),arm64)
GOARCH = arm64
else ifeq ($(shell uname -m),x86_64)
GOARCH = amd64
endif

# Detect OS
ifeq ($(OS),Windows_NT)
OS_DISTRIBUTION := Windows
else ifeq ($(shell uname -s),Darwin)
OS_DISTRIBUTION := macOS
else
OS_DISTRIBUTION := $(shell lsb_release -si)
endif


# Source code paths
SKYEYE_SOURCES = $(shell find . -type f -name '*.go')
SKYEYE_SOURCES += go.mod go.sum
SKYEYE_BIN = skyeye
SKYEYE_SCALER_BIN = skyeye-scaler

WHISPER_CPP_PATH = third_party/whisper.cpp
WHISPER_CPP_BUILD_DIR = $(WHISPER_CPP_PATH)/build_go
LIBWHISPER_PATH = $(WHISPER_CPP_BUILD_DIR)/src/libwhisper.a
WHISPER_H_PATH = $(WHISPER_CPP_PATH)/include/whisper.h
WHISPER_CPP_REPO = https://github.com/ggml-org/whisper.cpp.git
WHISPER_CPP_VERSION = v1.8.4
WHISPER_CPP_CMAKE_ARGS =

# Compiler variables and flags
GOBUILDVARS = GOARCH=$(GOARCH)
ABS_WHISPER_CPP_PATH = $(abspath $(WHISPER_CPP_PATH))
ABS_WHISPER_CPP_BUILD_DIR = $(abspath $(WHISPER_CPP_BUILD_DIR))
LIBRARY_PATHS = $(ABS_WHISPER_CPP_BUILD_DIR)/src:$(ABS_WHISPER_CPP_BUILD_DIR)/ggml/src
BUILD_VARS = CGO_ENABLED=1 \
  C_INCLUDE_PATH="$(ABS_WHISPER_CPP_PATH)/ggml/include:$(ABS_WHISPER_CPP_PATH)/include" \
  LIBRARY_PATH="$(LIBRARY_PATHS)"
BUILD_FLAGS = -tags nolibopusfile

# Populate --version from Git tag
ifeq ($(SKYEYE_VERSION),)
SKYEYE_VERSION=$(shell git describe --tags || echo devel)
endif
LDFLAGS= -X "main.Version=$(SKYEYE_VERSION)"

# macOS-specific settings
ifeq ($(OS_DISTRIBUTION),macOS)
# Use Homebrew LLVM/Clang for OpenMP support
CC=$(shell brew --prefix llvm)/bin/clang
CXX=$(shell brew --prefix llvm)/bin/clang++
BUILD_VARS += CC=$(CC) CXX=$(CXX)
LIBRARY_PATHS := $(LIBRARY_PATHS):$(ABS_WHISPER_CPP_BUILD_DIR)/ggml/src/ggml-metal:$(ABS_WHISPER_CPP_BUILD_DIR)/ggml/src/ggml-blas
# Link OpenMP runtime for ggml-cpu on macOS (Go bindings only specify -fopenmp on Linux)
BUILD_VARS += CGO_LDFLAGS=-fopenmp
WHISPER_CPP_CMAKE_ARGS = -DCMAKE_C_COMPILER=$(CC) -DCMAKE_CXX_COMPILER=$(CXX) \
  -DCMAKE_C_FLAGS=-Wno-elaborated-enum-base -DCMAKE_CXX_FLAGS=-Wno-elaborated-enum-base
endif

# Windows-specific settings
ifeq ($(OS_DISTRIBUTION),Windows)
# Compile EXE instead of ELF
SKYEYE_BIN = skyeye.exe
SKYEYE_SCALER_BIN = skyeye-scaler.exe
# Override Windows Go environment with MSYS2 UCRT64 Go environment
GO = /ucrt64/bin/go
GOBUILDVARS += GOROOT="/ucrt64/lib/go" GOPATH="/ucrt64"
# Use MSYS Makefiles generator for cmake on MSYS2
WHISPER_CPP_CMAKE_ARGS = -G "MSYS Makefiles"
# Static linking on Windows to avoid MSYS2 dependency at runtime
LIBRARIES = opus soxr
CFLAGS = $(shell pkg-config $(LIBRARIES) --cflags --static)
BUILD_VARS += CFLAGS='$(CFLAGS)'
EXTLDFLAGS = $(shell pkg-config $(LIBRARIES) --libs --static)
LDFLAGS += -linkmode external -extldflags "$(EXTLDFLAGS) -static"
endif

BUILD_VARS += LDFLAGS='$(LDFLAGS)'
BUILD_FLAGS += -ldflags '$(LDFLAGS)'
GO := $(GOBUILDVARS) $(GO)

.PHONY: default
default: $(SKYEYE_BIN)

.PHONY: install-msys2-dependencies
install-msys2-dependencies:
	pacman -Syu --needed \
	  git \
	  base-devel \
	  $(MINGW_PACKAGE_PREFIX)-cmake \
	  $(MINGW_PACKAGE_PREFIX)-toolchain \
	  $(MINGW_PACKAGE_PREFIX)-go \
	  $(MINGW_PACKAGE_PREFIX)-opus \
	  $(MINGW_PACKAGE_PREFIX)-libsoxr

.PHONY: install-arch-linux-dependencies
install-arch-linux-dependencies:
	sudo pacman -Syu \
	  git \
	  base-devel \
	  cmake \
	  go \
	  opus \
	  libsoxr

.PHONY: install-debian-dependencies
install-debian-dependencies:
	sudo apt-get update
	sudo apt-get install -y \
	  git \
	  build-essential \
	  cmake \
	  golang-go \
	  libopus-dev \
	  libopus0 \
	  libsoxr-dev \
	  libsoxr0

.PHONY: install-fedora-dependencies
install-fedora-dependencies:
	sudo dnf install -y \
	  git \
	  development-tools \
	  c-development \
	  cmake \
	  golang \
	  opus-devel \
	  opus \
	  soxr-devel \
	  sox

.PHONY: install-macos-dependencies
install-macos-dependencies:
	xcode-select --install || true
	brew install \
	  git \
	  cmake \
	  llvm \
	  pkg-config \
	  go \
	  libsoxr \
	  opus

.PHONY: download-whisper-%
download-whisper-%:
	curl -L -o $*.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/$*.bin

$(LIBWHISPER_PATH) $(WHISPER_H_PATH):
	if [ ! -f $(LIBWHISPER_PATH) -o ! -f $(WHISPER_H_PATH) ]; then \
		git -C "$(WHISPER_CPP_PATH)" checkout --quiet $(WHISPER_CPP_VERSION) || \
		git clone --depth 1 --branch $(WHISPER_CPP_VERSION) -c advice.detachedHead=false "$(WHISPER_CPP_REPO)" "$(WHISPER_CPP_PATH)" && \
		cmake -S "$(WHISPER_CPP_PATH)" -B "$(WHISPER_CPP_BUILD_DIR)" \
			-DCMAKE_BUILD_TYPE=Release \
			-DBUILD_SHARED_LIBS=OFF \
			$(WHISPER_CPP_CMAKE_ARGS) && \
		cmake --build "$(WHISPER_CPP_BUILD_DIR)" --target whisper && \
		for f in $$(find "$(WHISPER_CPP_BUILD_DIR)" -name '*.a' ! -name 'lib*'); do \
			mv "$$f" "$$(dirname $$f)/lib$$(basename $$f)"; \
		done; \
	fi

.PHONY: whisper
whisper: $(LIBWHISPER_PATH) $(WHISPER_H_PATH)

.PHONY: generate
generate:
	$(BUILD_VARS) $(GO) generate $(BUILD_FLAGS) ./...

$(SKYEYE_BIN): generate $(SKYEYE_SOURCES) $(LIBWHISPER_PATH) $(WHISPER_H_PATH)
	$(BUILD_VARS) $(GO) build $(BUILD_FLAGS) ./cmd/skyeye/

$(SKYEYE_SCALER_BIN): generate $(SKYEYE_SOURCES)
	$(BUILD_VARS) $(GO) build $(BUILD_FLAGS) ./cmd/skyeye-scaler/

.PHONY: run
run:
	$(BUILD_VARS) $(GO) run -race $(BUILD_FLAGS) ./cmd/skyeye/ $(ARGS)

.PHONY: test
test: generate
	$(BUILD_VARS) $(GO) tool gotestsum -- $(BUILD_FLAGS) $(TEST_FLAGS) ./...

.PHONY: benchmark-whisper
benchmark-whisper: whisper
	test -n "$(SKYEYE_WHISPER_MODEL)"  # Set SKYEYE_WHISPER_MODEL to the absolute path to the model's .bin file
	$(BUILD_VARS) $(GO) test -bench=. -run BenchmarkWhisperRecognizer ./pkg/recognizer

.PHONY: vet
vet: generate
	$(BUILD_VARS) $(GO) vet $(BUILD_FLAGS) ./...

# Note: Running golangci-lint from source like this is not recommended, see https://golangci-lint.run/welcome/install/#install-from-source
# However, this is the easiest way to set the required CGO variables for this project.
.PHONY: lint
lint: whisper generate
	$(BUILD_VARS) $(GO) tool golangci-lint run ./...


.PHONY: fix
fix: generate
	$(BUILD_VARS) $(GO) fix $(BUILD_FLAGS) ./...

.PHONY: format
format:
	find . -name '*.go' -exec gofmt -s -w {} ';'

.PHONY: mostlyclean
mostlyclean:
	rm -f "$(SKYEYE_BIN)" "$(SKYEYE_SCALER_BIN)"
	find . -type f -name 'mock_*.go' -delete

.PHONY: clean
clean: mostlyclean
	rm -rf "$(WHISPER_CPP_PATH)"
