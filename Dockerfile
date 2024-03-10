FROM golang:1.22 as builder
ARG whisper_cpp_version=v1.5.4
RUN apt-get update && apt-get install -y git libasound2-dev libopus-dev libsoxr-dev
WORKDIR /app
RUN git clone --depth 1 --branch ${whisper_cpp_version} -c advice.detachedHead=false https://github.com/ggerganov/whisper.cpp.git third_party/whisper.cpp
RUN make -C third_party/whisper.cpp/bindings/go whisper
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=1 C_INCLUDE_PATH=/app/third_party/whisper.cpp LIBRARY_PATH=/app/third_party/whisper.cpp go build -o skyeye -tags nolibopusfile ./cmd/skyeye
