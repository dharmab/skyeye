ARG SKYEYE_VERSION
FROM golang:1.23.3 AS builder
RUN apt-get update && apt-get install -y \
  git \
  make \
  lsb-release \
  gcc \
  libopus-dev \
  libsoxr-dev \
  libopenblas-openmp-dev \
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

FROM debian:bookworm-slim AS base
RUN apt-get update && apt-get install -y \
  libopus0 \
  libsoxr0 \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

FROM base AS skyeye
RUN apt-get update && apt-get install -y \
  libopenblas0-openmp \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /skyeye/skyeye /opt/skyeye/bin/skyeye
ENTRYPOINT ["/opt/skyeye/bin/skyeye"]

FROM base AS skyeye-scaler
COPY --from=builder /skyeye/skyeye-scaler /opt/skyeye/bin/skyeye-scaler
ENTRYPOINT ["/opt/skyeye/bin/skyeye-scaler"]