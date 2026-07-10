ARG SKYEYE_VERSION
FROM golang:1.26 AS builder
RUN apt-get update && apt-get install -y \
    git \
    make \
    cmake \
    lsb-release \
    gcc \
    libopus-dev \
    libsoxr-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /skyeye
COPY third_party third_party
COPY Makefile Makefile
RUN make whisper
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download -x
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN make skyeye
RUN make skyeye-scaler

FROM debian:trixie-slim AS base
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libopus0 \
    libsoxr0 \
    && rm -rf /var/lib/apt/lists/*

FROM base AS skyeye
COPY --from=builder /skyeye/skyeye /opt/skyeye/bin/skyeye
ENTRYPOINT ["/opt/skyeye/bin/skyeye"]

FROM base AS skyeye-scaler
COPY --from=builder /skyeye/skyeye-scaler /opt/skyeye/bin/skyeye-scaler
ENTRYPOINT ["/opt/skyeye/bin/skyeye-scaler"]

FROM builder AS builder-vulkan
RUN apt-get update && apt-get install -y \
    libvulkan-dev \
    glslc \
    spirv-headers \
    && rm -rf /var/lib/apt/lists/*
RUN make whisper GGML_VULKAN=1
RUN make skyeye-vulkan

FROM base AS skyeye-vulkan
# libvulkan1: Vulkan loader. mesa-vulkan-drivers: Vulkan ICDs for AMD/Intel GPUs
# passed in with --device /dev/dri. NVIDIA users don't need it: the NVIDIA
# container toolkit injects NVIDIA's own ICD. Unused ICDs are harmless.
RUN apt-get update && apt-get install -y \
    libvulkan1 \
    mesa-vulkan-drivers \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder-vulkan /skyeye/skyeye-vulkan /opt/skyeye/bin/skyeye
ENTRYPOINT ["/opt/skyeye/bin/skyeye"]
