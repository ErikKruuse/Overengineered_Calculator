# ---- Build stage ----
FROM golang:1.23-alpine AS builder

# Install git (required for Go modules using private repos)
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary statically (no CGO dependencies)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./server

# ---- Runtime stage ----
FROM alpine:latest

# Minimal CA certificates (for HTTPS clients)
RUN apk --no-cache add ca-certificates

# Copy compiled binary from builder
WORKDIR /root/
COPY --from=builder /app/server .

# Cloud Run and other container hosts set $PORT
ENV PORT=8080

# Expose the port (useful locally)
EXPOSE 8080

# Run the binary
CMD ["./server"]
