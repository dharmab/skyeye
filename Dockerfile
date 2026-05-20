ARG SKYEYE_VERSION
ARG WHISPER_CPP_BACKEND=cpu

FROM golang:1.26 AS builder
ARG WHISPER_CPP_BACKEND
RUN apt-get update && apt-get install -y \
    git \
    make \
    cmake \
    lsb-release \
    gcc \
    libopus-dev \
    libsoxr-dev \
    && if [ "$WHISPER_CPP_BACKEND" = "vulkan" ]; then \
      apt-get install -y libvulkan-dev glslc; \
    fi \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /skyeye
COPY third_party third_party
COPY Makefile Makefile
RUN make whisper WHISPER_CPP_BACKEND=$WHISPER_CPP_BACKEND
COPY go.mod go.sum ./
RUN go mod download -x
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN make skyeye WHISPER_CPP_BACKEND=$WHISPER_CPP_BACKEND
RUN make skyeye-scaler

FROM debian:bookworm-slim AS base
ARG WHISPER_CPP_BACKEND
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libopus0 \
    libsoxr0 \
    && if [ "$WHISPER_CPP_BACKEND" = "vulkan" ]; then \
      apt-get install -y libvulkan1 mesa-vulkan-drivers; \
    fi \
    && rm -rf /var/lib/apt/lists/*

FROM base AS skyeye
COPY --from=builder /skyeye/skyeye /opt/skyeye/bin/skyeye
ENTRYPOINT ["/opt/skyeye/bin/skyeye"]

FROM base AS skyeye-scaler
COPY --from=builder /skyeye/skyeye-scaler /opt/skyeye/bin/skyeye-scaler
ENTRYPOINT ["/opt/skyeye/bin/skyeye-scaler"]
