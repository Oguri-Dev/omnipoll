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

# Ensure web/dist is present (will be copied before build)
RUN if [ ! -d "web/dist" ]; then echo "ERROR: web/dist not found. Run 'cp -r frontend/dist backend/web/' before building"; exit 1; fi

# Build binary from cmd/omnipoll (which has embed directive)
WORKDIR /app/cmd/omnipoll
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/omnipoll .

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
RUN mkdir -p /app/data

# Expose admin panel port
EXPOSE 8080

# Run as non-root user
RUN adduser -D -u 1000 omnipoll
USER omnipoll

ENTRYPOINT ["./omnipoll"]
