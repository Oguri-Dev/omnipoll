# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Ensure web/dist is present
RUN if [ ! -d "web/dist" ]; then echo "ERROR: web/dist not found. Run 'cp -r frontend/dist backend/web/' before building"; exit 1; fi

# Build binary (compile from /app so embed can find web/dist)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o omnipoll ./cmd/omnipoll

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/omnipoll .

# Copy frontend static files
COPY --from=builder /app/web/dist ./web/dist

# Create data directory for watermark persistence
RUN mkdir -p /app/data && chmod 777 /app/data

# Expose admin panel port
EXPOSE 8080

# Run as root to avoid permission issues (or remove USER line)
# USER omnipoll

ENTRYPOINT ["./omnipoll"]
