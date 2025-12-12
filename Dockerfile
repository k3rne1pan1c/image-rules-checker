# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o imgcheck \
    ./cmd/imgcheck

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS if needed
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/imgcheck /app/imgcheck
COPY --from=builder /build/config /app/config

# Make binary executable
RUN chmod +x /app/imgcheck

# Default command
ENTRYPOINT ["/app/imgcheck"]

