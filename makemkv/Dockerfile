FROM ubuntu:24.04

ARG MAKEMKV_VERSION=1.18.3

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    libc6-dev \
    libssl-dev \
    libexpat1-dev \
    libavcodec-dev \
    libgl1-mesa-dev \
    qtbase5-dev \
    zlib1g-dev \
    wget \
    ca-certificates \
    tar \
    ffmpeg \
    mkvtoolnix \
    less \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

RUN wget "https://www.makemkv.com/download/makemkv-oss-${MAKEMKV_VERSION}.tar.gz" && \
    wget "https://www.makemkv.com/download/makemkv-bin-${MAKEMKV_VERSION}.tar.gz" && \
    tar -xzf "makemkv-oss-${MAKEMKV_VERSION}.tar.gz" && \
    tar -xzf "makemkv-bin-${MAKEMKV_VERSION}.tar.gz"

RUN cd "/tmp/makemkv-oss-${MAKEMKV_VERSION}" && \
    ./configure && \
    make -j"$(nproc)" && \
    make install

RUN cd "/tmp/makemkv-bin-${MAKEMKV_VERSION}" && \
    make -j"$(nproc)" && \
    printf "yes\n" | make install

ENV HOME=/config

WORKDIR /work

ENTRYPOINT ["makemkvcon"]