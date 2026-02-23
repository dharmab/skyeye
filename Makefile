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

# Compiler variables and flags
GOBUILDVARS = GOARCH=$(GOARCH)
BUILD_VARS = CGO_ENABLED=1
BUILD_FLAGS = -tags nolibopusfile

# Populate --version from Git tag
ifeq ($(SKYEYE_VERSION),)
SKYEYE_VERSION=$(shell git describe --tags || echo devel)
endif
LDFLAGS= -X "main.Version=$(SKYEYE_VERSION)"

# Windows-specific settings
ifeq ($(OS_DISTRIBUTION),Windows)
# Compile EXE instead of ELF
SKYEYE_BIN = skyeye.exe
SKYEYE_SCALER_BIN = skyeye-scaler.exe
# Override Windows Go environment with MSYS2 UCRT64 Go environment
GO = /ucrt64/bin/go
GOBUILDVARS += GOROOT="/ucrt64/lib/go" GOPATH="/ucrt64"
# On Windows, we statically link opus and soxr so users don't need to install them.
LIBRARIES = opus soxr
CFLAGS = $(shell pkg-config $(LIBRARIES) --cflags --static)
BUILD_VARS += CFLAGS='$(CFLAGS)'
EXTLDFLAGS = -Wl,-Bstatic $(shell pkg-config $(LIBRARIES) --libs --static) -Wl,-Bdynamic
LDFLAGS += -linkmode external -extldflags "$(EXTLDFLAGS)"
# On Windows, we copy the ONNX Runtime DLLs so we can package them with the binary during distribution.
# The module version is read directly from go.mod to avoid invoking Go (which would trigger
# toolchain delegation to a Windows-native binary that misinterprets MSYS2 POSIX paths).
SHERPA_VERSION := $(shell grep 'k2-fsa/sherpa-onnx-go-windows' go.mod | awk '{print $$2}')
SHERPA_DLL_DIR := /ucrt64/pkg/mod/github.com/k2-fsa/sherpa-onnx-go-windows@$(SHERPA_VERSION)/lib/x86_64-pc-windows-gnu
SHERPA_DLLS = sherpa-onnx-c-api.dll onnxruntime.dll sherpa-onnx-cxx-api.dll
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
	  build-essential \
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
	  pkg-config \
	  go \
	  libsoxr \
	  opus

.PHONY: generate
generate:
ifeq ($(OS_DISTRIBUTION),Windows)
	SHERPA_LIB="$(SHERPA_DLL_DIR)" $(BUILD_VARS) $(GO) generate $(BUILD_FLAGS) ./...
else
	$(BUILD_VARS) $(GO) generate $(BUILD_FLAGS) ./...
endif

$(SKYEYE_BIN): generate $(SKYEYE_SOURCES)
	$(BUILD_VARS) $(GO) build $(BUILD_FLAGS) ./cmd/skyeye/
ifeq ($(OS_DISTRIBUTION),Windows)
	cp $(addprefix $(SHERPA_DLL_DIR)/,$(SHERPA_DLLS)) .
endif

$(SKYEYE_SCALER_BIN): generate $(SKYEYE_SOURCES)
	$(BUILD_VARS) $(GO) build $(BUILD_FLAGS) ./cmd/skyeye-scaler/

.PHONY: download-models
download-models:
	CGO_ENABLED=0 $(GO) run ./cmd/download-models $(ARGS)

.PHONY: run
run:
	$(BUILD_VARS) $(GO) run -race $(BUILD_FLAGS) ./cmd/skyeye/ $(ARGS)

.PHONY: test
test: generate
	$(BUILD_VARS) $(GO) tool gotestsum -- $(BUILD_FLAGS) $(TEST_FLAGS) ./...

.PHONY: benchmark-parakeet
benchmark-parakeet:
	$(BUILD_VARS) $(GO) test -bench=. -run BenchmarkParakeetRecognizer ./pkg/recognizer/parakeet

.PHONY: vet
vet: generate
	$(BUILD_VARS) $(GO) vet $(BUILD_FLAGS) ./...

# Note: Running golangci-lint from source like this is not recommended, see https://golangci-lint.run/welcome/install/#install-from-source
# However, this is the easiest way to set the required CGO variables for this project.
.PHONY: lint
lint: generate
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
ifeq ($(OS_DISTRIBUTION),Windows)
	rm -f $(SHERPA_DLLS)
endif
	find . -type f -name 'mock_*.go' -delete

.PHONY: clean
clean: mostlyclean
