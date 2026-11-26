# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w -X github.com/d0whc3r/go-s3/internal/version.Version=${VERSION}" \
    -o /out/gos3 .

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source="https://github.com/d0whc3r/go-s3" \
      org.opencontainers.image.description="S3-compatible backup and retention CLI" \
      org.opencontainers.image.licenses="Apache-2.0"

COPY --from=builder /out/gos3 /usr/local/bin/gos3

ENTRYPOINT ["/usr/local/bin/gos3"]
