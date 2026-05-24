FROM golang:1.26-alpine

RUN apk add --no-cache \
    ffmpeg \
    tesseract-ocr \
    tesseract-ocr-dev \
    tesseract-ocr-data-eng \
    leptonica-dev \
    build-base \
    pkgconf \
    ca-certificates

WORKDIR /app

COPY go.* ./

RUN go mod download || true

COPY . .

ENV CGO_ENABLED=1

RUN go build -o /usr/local/bin/nensaysubs .

ENV TESSDATA_PREFIX=/usr/share/tessdata

ENTRYPOINT ["/bin/sh", "-c"]
CMD ["tail -f /dev/null"]