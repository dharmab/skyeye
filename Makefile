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
WHISPER_CPP_BACKEND ?= cpu
WHISPER_CPP_BUILD_DIR_SUFFIX =
ifneq ($(WHISPER_CPP_BACKEND),cpu)
  WHISPER_CPP_BUILD_DIR_SUFFIX = _$(WHISPER_CPP_BACKEND)
endif
WHISPER_CPP_BUILD_DIR = $(WHISPER_CPP_PATH)/build_go$(WHISPER_CPP_BUILD_DIR_SUFFIX)
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
CGO_LDFLAGS =
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
CGO_LDFLAGS += -fopenmp
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
LDFLAGS += -linkmode external -extldflags "$(EXTLDFLAGS) -lgomp -static"
endif

# Vulkan backend settings
ifeq ($(WHISPER_CPP_BACKEND),vulkan)
  WHISPER_CPP_CMAKE_ARGS += -DGGML_VULKAN=ON
  LIBRARY_PATHS := $(LIBRARY_PATHS):$(ABS_WHISPER_CPP_BUILD_DIR)/ggml/src/ggml-vulkan
  BUILD_VARS := $(filter-out LIBRARY_PATH=%,$(BUILD_VARS)) LIBRARY_PATH="$(LIBRARY_PATHS)"
  ifeq ($(OS_DISTRIBUTION),Windows)
    CGO_LDFLAGS += -lggml-vulkan -lvulkan-1
  else
    CGO_LDFLAGS += -lggml-vulkan -lvulkan
  endif
endif

ifneq ($(strip $(CGO_LDFLAGS)),)
BUILD_VARS += CGO_LDFLAGS='$(CGO_LDFLAGS)'
endif
BUILD_VARS += LDFLAGS='$(LDFLAGS)'
BUILD_FLAGS += -ldflags '$(LDFLAGS)'
GO := $(GOBUILDVARS) $(GO)

# CI distribution variables
DIST_BACKEND_SUFFIX =
ifneq ($(WHISPER_CPP_BACKEND),cpu)
  DIST_BACKEND_SUFFIX = -$(WHISPER_CPP_BACKEND)
endif
ifeq ($(OS_DISTRIBUTION),macOS)
  DIST_OS = macos
else ifeq ($(OS_DISTRIBUTION),Windows)
  DIST_OS = windows
else
  DIST_OS = linux
endif
DIST_NAME = skyeye-$(DIST_OS)-$(GOARCH)$(DIST_BACKEND_SUFFIX)
DIST_DIR = dist/$(DIST_NAME)

.PHONY: default
default: $(SKYEYE_BIN)

.PHONY: whisper-vulkan skyeye-vulkan
whisper-vulkan:
	$(MAKE) WHISPER_CPP_BACKEND=vulkan whisper
skyeye-vulkan:
	$(MAKE) WHISPER_CPP_BACKEND=vulkan $(SKYEYE_BIN)

.PHONY: install-msys2-dependencies
install-msys2-dependencies:
	pacman -Syu --needed \
	  git \
	  base-devel \
	  $(MINGW_PACKAGE_PREFIX)-cmake \
	  $(MINGW_PACKAGE_PREFIX)-toolchain \
	  $(MINGW_PACKAGE_PREFIX)-go \
	  $(MINGW_PACKAGE_PREFIX)-opus \
	  $(MINGW_PACKAGE_PREFIX)-libsoxr \
	  $(MINGW_PACKAGE_PREFIX)-vulkan-headers \
	  $(MINGW_PACKAGE_PREFIX)-vulkan-loader \
	  $(MINGW_PACKAGE_PREFIX)-shaderc \
	  $(MINGW_PACKAGE_PREFIX)-spirv-tools

.PHONY: install-arch-linux-dependencies
install-arch-linux-dependencies:
	sudo pacman -Syu \
	  git \
	  base-devel \
	  cmake \
	  go \
	  opus \
	  libsoxr \
	  vulkan-headers \
	  vulkan-icd-loader \
	  shaderc

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
	  libsoxr0 \
	  libvulkan-dev \
	  glslc

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
	  sox \
	  vulkan-headers \
	  vulkan-loader-devel \
	  glslc

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

.PHONY: dist-linux dist-macos
dist-linux dist-macos: $(SKYEYE_BIN) $(SKYEYE_SCALER_BIN)
	mkdir -p $(DIST_DIR)/docs/
	cp $(SKYEYE_BIN) $(DIST_DIR)/$(SKYEYE_BIN)
	cp $(SKYEYE_SCALER_BIN) $(DIST_DIR)/$(SKYEYE_SCALER_BIN)
	chmod +x $(DIST_DIR)/$(SKYEYE_BIN)
	chmod +x $(DIST_DIR)/$(SKYEYE_SCALER_BIN)
	cp README.md $(DIST_DIR)/README.md
	cp LICENSE $(DIST_DIR)/LICENSE
	cp config.yaml $(DIST_DIR)/config.yaml
	cp docs/*.md $(DIST_DIR)/docs/
	tar -czf dist/$(DIST_NAME).tar.gz -C dist $(DIST_NAME)

.PHONY: dist-linux-vulkan
dist-linux-vulkan:
	$(MAKE) WHISPER_CPP_BACKEND=vulkan dist-linux

WINSW_VERSION = 2.12.0
WINSW_SHA256 = 05b82d46ad331cc16bdc00de5c6332c1ef818df8ceefcd49c726553209b3a0da

winsw.exe:
	curl -fsL https://github.com/winsw/winsw/releases/download/v$(WINSW_VERSION)/WinSW-x64.exe -o winsw.exe
	echo "$(WINSW_SHA256)  winsw.exe" | sha256sum -c - || \
		(echo "ERROR: winsw.exe hash verification failed - expected SHA256 $(WINSW_SHA256)" && rm -f winsw.exe && exit 1)

.PHONY: dist-windows
dist-windows: $(SKYEYE_BIN) $(SKYEYE_SCALER_BIN) winsw.exe
	mkdir -p $(DIST_DIR)/docs/
	cp $(SKYEYE_BIN) $(DIST_DIR)/$(SKYEYE_BIN)
	cp $(SKYEYE_SCALER_BIN) $(DIST_DIR)/$(SKYEYE_SCALER_BIN)
	cp README.md $(DIST_DIR)/README.md
	cp LICENSE $(DIST_DIR)/LICENSE
	cp config.yaml $(DIST_DIR)/config.yaml
	cp docs/*.md $(DIST_DIR)/docs/
	cp winsw.exe $(DIST_DIR)/skyeye-service.exe
	cp init/winsw/skyeye-service.yml $(DIST_DIR)/skyeye-service.yml
	cp winsw.exe $(DIST_DIR)/skyeye-scaler-service.exe
	cp init/winsw/skyeye-scaler-service.yml $(DIST_DIR)/skyeye-scaler-service.yml
	cd dist && zip -r $(DIST_NAME).zip $(DIST_NAME)

.PHONY: dist-windows-vulkan
dist-windows-vulkan:
	$(MAKE) WHISPER_CPP_BACKEND=vulkan dist-windows

.PHONY: ci-lint
ci-lint: lint vet fix format
	$(GO) mod tidy
	git diff --exit-code

.PHONY: mostlyclean
mostlyclean:
	rm -f "$(SKYEYE_BIN)" "$(SKYEYE_SCALER_BIN)"
	rm -rf dist/
	find . -type f -name 'mock_*.go' -delete

.PHONY: clean
clean: mostlyclean
	rm -rf "$(WHISPER_CPP_PATH)"
	rm -f winsw.exe
