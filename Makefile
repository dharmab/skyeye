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
LIBWHISPER_PATH = $(WHISPER_CPP_PATH)/libwhisper.a
WHISPER_H_PATH = $(WHISPER_CPP_PATH)/include/whisper.h
WHISPER_CPP_REPO = https://github.com/dharmab/whisper.cpp.git
WHISPER_CPP_VERSION = v1.7.2-windows-fix
WHISPER_CPP_BUILD_ENV =

# Compiler variables and flags
GOBUILDVARS = GOARCH=$(GOARCH)
ABS_WHISPER_CPP_PATH = $(abspath $(WHISPER_CPP_PATH))
BUILD_VARS = CGO_ENABLED=1 \
  C_INCLUDE_PATH="$(ABS_WHISPER_CPP_PATH)/ggml/include:$(ABS_WHISPER_CPP_PATH)/include" \
  LIBRARY_PATH="$(ABS_WHISPER_CPP_PATH)"
BUILD_FLAGS = -tags nolibopusfile

# Populate --version from Git tag
ifeq ($(SKYEYE_VERSION),)
SKYEYE_VERSION=$(shell git describe --tags || echo devel)
endif
LDFLAGS= -X "main.Version=$(SKYEYE_VERSION) -fopenmp"

# macOS-specific settings
ifeq ($(OS_DISTRIBUTION),macOS)
# Use Homebrew LLVM/Clang for OpenMP support
CC=$(shell brew --prefix llvm)/bin/clang
CXX=$(shell brew --prefix llvm)/bin/clang++
BUILD_VARS += CC=$(CC) CXX=$(CXX)
# Enable GPU acceleration
WHISPER_CPP_BUILD_ENV = GGML_METAL=1
endif

# Windows-specific settings
ifeq ($(OS_DISTRIBUTION),Windows)
# Compile EXE instead of ELF
SKYEYE_BIN = skyeye.exe
SKYEYE_SCALER_BIN = skyeye-scaler.exe
# Override Windows Go environment with MSYS2 UCRT64 Go environment
GO = /ucrt64/bin/go
GOBUILDVARS += GOROOT="/ucrt64/lib/go" GOPATH="/ucrt64"
# Static linking on Windows to avoid MSYS2 dependency at runtime
LIBRARIES = opus soxr
CFLAGS = $(pkg-config $(LIBRARIES) --cflags --static)
BUILD_VARS += CFLAGS=$(CFLAGS)
EXTLDFLAGS = $(pkg-config $(LIBRARIES) --libs --static)
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
	  $(MINGW_PACKAGE_PREFIX)-toolchain \
	  $(MINGW_PACKAGE_PREFIX)-go \
	  $(MINGW_PACKAGE_PREFIX)-opus \
	  $(MINGW_PACKAGE_PREFIX)-libsoxr

.PHONY: install-arch-linux-dependencies
install-arch-linux-dependencies:
	sudo pacman -Syu \
	  git \
	  base-devel \
	  go \
	  opus \
	  libsoxr

.PHONY: install-debian-dependencies
install-debian-dependencies:
	sudo apt-get update
	sudo apt-get install -y \
	  git \
	  golang-go \
	  libopus-dev \
	  libopus0 \
	  libsoxr-dev \
	  libsoxr0

.PHONY: install-macos-dependencies
install-macos-dependencies:
	xcode-select --install || true
	brew install \
	  git \
	  go \
	  libomp \
	  libsoxr \
	  llvm \
	  opus

$(LIBWHISPER_PATH) $(WHISPER_H_PATH):
	if [ ! -f $(LIBWHISPER_PATH) -o ! -f $(WHISPER_H_PATH) ]; then git -C "$(WHISPER_CPP_PATH)" checkout --quiet $(WHISPER_CPP_VERSION) || git clone --depth 1 --branch $(WHISPER_CPP_VERSION) -c advice.detachedHead=false "$(WHISPER_CPP_REPO)" "$(WHISPER_CPP_PATH)" && $(WHISPER_CPP_BUILD_ENV) make -C $(WHISPER_CPP_PATH)/bindings/go whisper; fi
	if [ -f third_party/whisper.cpp/whisper.a ] && [ ! -f third_party/whisper.cpp/libwhisper.a ]; then cp third_party/whisper.cpp/whisper.a third_party/whisper.cpp/libwhisper.a; fi

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
# However, this is the easiest way set the required CGO variables for this project.
.PHONY: lint
lint: whisper generate
	$(BUILD_VARS) $(GO) tool golangci-lint run ./...

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
