# ---- Build stage ----
FROM golang:1.25.3-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src

# If you don't have go.sum, only copy go.mod first
COPY go.mod ./
RUN go mod download

# Copy the rest of your source
COPY . .

# Build to a path that won't clash with a "server/" dir in your repo
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/calcserver ./server

# ---- Runtime stage ----
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy just the compiled binary
COPY --from=builder /out/calcserver /usr/local/bin/calcserver

ENV PORT=8080
EXPOSE 8080

# Execute the binary (now clearly not a directory)
CMD ["calcserver"]
