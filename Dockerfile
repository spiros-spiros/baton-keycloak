# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install git
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o baton-keycloak ./cmd/baton-keycloak

# Final stage
FROM alpine:latest

# Create a non-root user first
RUN adduser -D -g '' appuser

# Create app directory and set permissions
RUN mkdir -p /app && chown -R appuser:appuser /app

WORKDIR /app

# Copy the binary from builder
COPY --from=builder --chown=appuser:appuser /app/baton-keycloak /app/

USER appuser

# Run the application
ENTRYPOINT ["/app/baton-keycloak"]
