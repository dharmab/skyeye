.PHONY: default
default: build

.PHONY: install-dependencies
install-dependencies:
	pacman -Syu --needed git base-devel $(MINGW_PACKAGE_PREFIX)-toolchain $(MINGW_PACKAGE_PREFIX)-go $(MINGW_PACKAGE_PREFIX)-opus

WHISPER_CPP_PATH = third_party/whisper.cpp
WHISPER_CPP_VERSION = v1.5.4

.PHONY: $(WHISPER_CPP_PATH)
$(WHISPER_CPP_PATH):
	git -C "$(WHISPER_CPP_PATH)" checkout --quiet $(WHISPER_CPP_VERSION) || git clone --depth 1 --branch $(WHISPER_CPP_VERSION) -c advice.detachedHead=false https://github.com/ggerganov/whisper.cpp.git "$(WHISPER_CPP_PATH)"

$(WHISPER_CPP_PATH)/libwhisper.a $(WHISPER_CPP_PATH)/whisper.h &: $(WHISPER_CPP_PATH)
	make -C $(WHISPER_CPP_PATH)/bindings/go whisper

.PHONY: whisper
whisper: $(WHISPER_CPP_PATH)/libwhisper.a $(WHISPER_CPP_PATH)/whisper.h

SKYEYE_PATH = $(shell pwd)
SKYEYE_SOURCES = $(shell find . -type f -name '*.go')
SKYEYE_SOURCES += go.mod go.sum
SKYEYE_EXE = skyeye.exe
MSYS2_GOPATH = /mingw64
MSYS2_GOROOT = /mingw64/lib/go
MSYS2_GO = /mingw64/bin/go
BUILD_VARS = GOROOT="$(MSYS2_GOROOT)" GOPATH="$(MSYS2_GOPATH)" CGO_ENABLED=1 C_INCLUDE_PATH="$(SKYEYE_PATH)/$(WHISPER_CPP_PATH)" LIBRARY_PATH="$(SKYEYE_PATH)/$(WHISPER_CPP_PATH)"
BUILD_TAGS = -tags nolibopusfile

$(SKYEYE_EXE): $(SKYEYE_SOURCES) $(WHISPER_CPP_PATH)/libwhisper.a $(WHISPER_CPP_PATH)/whisper.h
	$(BUILD_VARS) $(MSYS2_GO) build $(BUILD_TAGS) ./cmd/skyeye/

.PHONY: build
build: $(SKYEYE_EXE)

.PHONY: test
test:
	go test $(BUILD_TAGS) ./...

.PHONY: mostlyclean
mostlyclean:
	rm -f "$(SKYEYE_EXE)"

.PHONY: clean
clean: mostlyclean
	rm -rf "$(WHISPER_CPP_PATH)"