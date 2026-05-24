FROM golang:1.21-alpine

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data \
    tesseract-dev \
    leptonica-dev \
    build-base \
    ca-certificates

WORKDIR /app

COPY go.* ./

RUN go mod download || true

COPY . .

RUN go build -o /usr/local/bin/nensaysubs .

ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr-4.1.1/tessdata

ENTRYPOINT ["/bin/sh", "-c"]
CMD ["tail -f /dev/null"]
