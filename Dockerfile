# Build stage
FROM golang:1.23-alpine AS builder

ARG VERSION=v0.1.0-alpha.1
ARG COMMIT=none
ARG DATE=unknown

WORKDIR /build

# Install ca-certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o romm-exporter .

# Final stage
FROM alpine:3.20

# Install runtime dependencies for SSL and Timezone support
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /

# Copy binary from builder
COPY --from=builder /build/romm-exporter /romm-exporter

# Default environment variables
ENV LISTEN_ADDR=":8585" \
    PUID=1000 \
    PGID=1000 \
    SCRAPE_TIMEOUT="10s"

EXPOSE 8585

ENTRYPOINT ["/romm-exporter"]
