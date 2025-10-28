# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    gpgme-dev \
    libassuan-dev \
    btrfs-progs-dev \
    device-mapper-dev \
    pkgconf

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    gpgme \
    libassuan \
    btrfs-progs-libs \
    device-mapper-libs

# Create non-root user
RUN addgroup -g 1000 zot && \
    adduser -D -u 1000 -G zot zot

# Create data directories
RUN mkdir -p /var/lib/zot/artifacts /var/lib/zot/metadata && \
    chown -R zot:zot /var/lib/zot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/zot-artifact-store /app/zot-artifact-store

# Switch to non-root user
USER zot

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/zot-artifact-store", "health", "--url", "http://localhost:8080/health"]

# Set entrypoint
ENTRYPOINT ["/app/zot-artifact-store"]
CMD ["serve"]
