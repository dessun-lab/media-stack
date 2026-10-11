FROM python:3.12-alpine

RUN apk add --no-cache \
    ffmpeg \
    curl \
    bash

RUN pip install --no-cache-dir yt-dlp

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app

WORKDIR /app
USER app

ENTRYPOINT ["yt-dlp"]