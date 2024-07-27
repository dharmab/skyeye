SKYEYE_EXE = skyeye.exe
SKYEYE_ELF = skyeye

.PHONY: default
ifeq ($(OS),Windows_NT)
default: $(SKYEYE_EXE)
else
default: $(SKYEYE_ELF)
endif

.PHONY: install-msys2-dependencies
install-msys2-dependencies:
	pacman -Syu --needed git base-devel $(MINGW_PACKAGE_PREFIX)-toolchain $(MINGW_PACKAGE_PREFIX)-go $(MINGW_PACKAGE_PREFIX)-opus $(MINGW_PACKAGE_PREFIX)-libsoxr

.PHONY: install-arch-linux-dependencies
install-arch-linux-dependencies:
	sudo pacman -Syu git base-devel alsa-lib go opus soxr

.PHONY: install-debian-dependencies
install-debian-dependencies:
	sudo apt-get update
	sudo apt-get install -y git libasound2-dev libopus-dev libsoxr-dev

WHISPER_CPP_PATH = third_party/whisper.cpp
LIBWHISPER_PATH = $(WHISPER_CPP_PATH)/libwhisper.a
WHISPER_H_PATH = $(WHISPER_CPP_PATH)/whisper.h
WHISPER_CPP_VERSION = v1.6.2

.PHONY: $(WHISPER_CPP_PATH)
$(WHISPER_CPP_PATH):
	git -C "$(WHISPER_CPP_PATH)" checkout --quiet $(WHISPER_CPP_VERSION) || git clone --depth 1 --branch $(WHISPER_CPP_VERSION) -c advice.detachedHead=false https://github.com/ggerganov/whisper.cpp.git "$(WHISPER_CPP_PATH)"

$(LIBWHISPER_PATH) $(WHISPER_H_PATH) &: $(WHISPER_CPP_PATH)
	make -C $(WHISPER_CPP_PATH)/bindings/go whisper

.PHONY: whisper
whisper: $(LIBWHISPER_PATH) $(WHISPER_H_PATH)

SKYEYE_PATH = $(shell pwd)
SKYEYE_SOURCES = $(shell find . -type f -name '*.go')
SKYEYE_SOURCES += go.mod go.sum

BUILD_VARS = CGO_ENABLED=1 C_INCLUDE_PATH="$(SKYEYE_PATH)/$(WHISPER_CPP_PATH)" LIBRARY_PATH="$(SKYEYE_PATH)/$(WHISPER_CPP_PATH)"
BUILD_TAGS = -tags nolibopusfile

MSYS2_GOPATH = /mingw64
MSYS2_GOROOT = /mingw64/lib/go
MSYS2_GO = /mingw64/bin/go

ifeq ($(OS),Windows_NT)
GO = $(MSYS2_GO)
else
GO = go
endif

.PHONY: generate
generate:
	$(BUILD_VARS) $(GO) generate $(BUILD_TAGS) ./...

$(SKYEYE_EXE): generate $(SKYEYE_SOURCES) $(LIBWHISPER_PATH) $(WHISPER_H_PATH)
	GOROOT="$(MSYS2_GOROOT)" GOPATH="$(MSYS2_GOPATH)" $(BUILD_VARS) $(GO) build $(BUILD_TAGS) ./cmd/skyeye/

$(SKYEYE_ELF): generate $(SKYEYE_SOURCES) $(LIBWHISPER_PATH) $(WHISPER_H_PATH)
	$(BUILD_VARS) $(GO) build $(BUILD_TAGS) ./cmd/skyeye/

.PHONY: test
test: generate
	$(BUILD_VARS) $(GO) test $(BUILD_TAGS) ./...

.PHONY: mostlyclean
mostlyclean:
	rm -f "$(SKYEYE_EXE)" "$(SKYEYE_ELF)"
	find . -type f -name 'mock_*.go' -delete

.PHONY: clean
clean: mostlyclean
	rm -rf "$(WHISPER_CPP_PATH)"