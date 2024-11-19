# Build
FROM golang:1.21.0-alpine AS builder
RUN apk add build-base
WORKDIR /app
COPY . /app
RUN go mod download
RUN go build -ldflags="-s -w" -trimpath cmd/acmeGoBaidu/acmeGoBaidu.go

FROM alpine:3.18.3
WORKDIR /app
COPY --from=builder /app/acmeGoBaidu /usr/local/bin/acmeGoBaidu


ENTRYPOINT ["acmeGoBaidu"]
