# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

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

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/baton-keycloak .

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Set environment variables
ENV KEYCLOAK_SERVER_URL=""
ENV KEYCLOAK_REALM=""
ENV KEYCLOAK_CLIENT_ID=""
ENV KEYCLOAK_CLIENT_SECRET=""

# Run the application
ENTRYPOINT ["./baton-keycloak"]
