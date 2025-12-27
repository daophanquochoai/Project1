# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/http-server ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget netcat-openbsd

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/http-server .

# Copy config files
COPY --from=builder /app/config .

EXPOSE 8005 9005

CMD ["./http-server"]