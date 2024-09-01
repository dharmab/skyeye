ARG SKYEYE_VERSION
FROM golang:1.23.0 AS builder
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

FROM debian:bookworm-slim AS skyeye
RUN apt-get update && apt-get install -y \
  libopus0 \
  libsoxr0 \
  libopenblas0-openmp \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /skyeye/skyeye /opt/skyeye/bin/skyeye
ENTRYPOINT ["/opt/skyeye/bin/skyeye"]
