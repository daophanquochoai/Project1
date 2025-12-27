# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/http-server ./cmd/http

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget
RUN cp /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && echo "Asia/Ho_Chi_Minh" > /etc/timezone

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/http-server .

# Copy config files
COPY --from=builder /app/config ./config

EXPOSE 8010

# Default command runs HTTP server, can be overridden in docker-compose
CMD ["./http-server"]

