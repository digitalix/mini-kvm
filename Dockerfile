# syntax = docker/dockerfile:1

FROM debian:bookworm as build

RUN apt-get update && \
    apt-get install -y \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev \
    libgstreamer-plugins-bad1.0-dev \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-libav \
    gstreamer1.0-tools \
    gstreamer1.0-alsa \
    gstreamer1.0-pulseaudio \
    pkg-config \
    gcc \
    g++ \
    make \
    wget

ARG GO_VER=1.25.2
ARG GO_ARCH="arm64"
RUN wget --no-check-certificate https://dl.google.com/go/go${GO_VER}.linux-${GO_ARCH}.tar.gz && \
    tar -xvf go${GO_VER}.linux-${GO_ARCH}.tar.gz && rm -r go${GO_VER}.linux-${GO_ARCH}.tar.gz && \
    mv go /usr/local

ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/go

COPY . /app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache \
    cd /app && \
    GODEBUG=cgocheck=0 go build -v -ldflags="-w -s" -o /app/service main.go

FROM scratch

COPY --from=build /app/service /