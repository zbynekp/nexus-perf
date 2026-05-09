# Multi-stage build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o nexus-perf

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/nexus-perf /app/

# Create non-root user for security
RUN addgroup -g 1000 nexus-perf && \
    adduser -D -u 1000 -G nexus-perf nexus-perf

USER nexus-perf

# Set entrypoint
ENTRYPOINT ["/app/nexus-perf"]
CMD ["--help"]

# Labels for metadata
LABEL maintainer="Performance Testing Team"
LABEL description="Nexus Repository Performance Testing Tool"
LABEL version="1.0.0"
