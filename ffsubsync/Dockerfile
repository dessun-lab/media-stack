FROM python:3.12-slim

RUN apt update && apt install -y ffmpeg \
    && pip install --no-cache-dir ffsubsync \
    && apt clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work

ENTRYPOINT ["ffsubsync"]