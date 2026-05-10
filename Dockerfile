# syntax=docker/dockerfile:1
# Multi-stage build: compile on golang:1.25-alpine (Alpine 3.21), run on alpine:3.21.
FROM golang:1.25-alpine AS builder

# git is needed by go mod download for VCS-based dependencies.
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_TIME=unknown
# TARGETARCH is injected by Docker BuildKit when building multi-platform images
# (docker buildx build --platform linux/amd64,linux/arm64).
# Falls back to amd64 for plain docker build.
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags "-X github.com/company/nexus-perf/cmd.appVersion=${VERSION} \
              -X github.com/company/nexus-perf/cmd.buildTime=${BUILD_TIME}" \
    -o nexus-perf .

# Runtime stage — keep Alpine version in sync with the builder base image.
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

# Create non-root user before copying files so ownership can be set in one step.
RUN addgroup -g 1000 nexus-perf && \
    adduser  -D -u 1000 -G nexus-perf nexus-perf

WORKDIR /app

COPY --from=builder --chown=nexus-perf:nexus-perf /app/nexus-perf /app/

USER nexus-perf

ENTRYPOINT ["/app/nexus-perf"]
CMD ["--help"]

ARG VERSION=dev
LABEL org.opencontainers.image.authors="Performance Testing Team"
LABEL org.opencontainers.image.description="Nexus Repository Performance Testing Tool"
LABEL org.opencontainers.image.version="${VERSION}"
