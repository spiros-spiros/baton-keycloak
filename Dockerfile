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

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/baton-keycloak .

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Set environment variables
ENV api_url="https://auth.dev.wcs.api.weaviate.io"
ENV realm="master"
ENV client="conductor-one-spiros"
ENV clientsecret="rTO2fzOydikCwZu8bdbYScoWlt4urPbZ"
ENV BATON_CLIENT_ID="aggressive-gorgon-76791@weaviate.conductor.one/ccc"
ENV BATON_CLIENT_SECRET="secret-token:conductorone.com:v1:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IldBTmFWaHphLXRLZ3RjV2NRTzhsRkdFbms4RUFVbWhRLTZmUXJBWFRUbkEiLCJkIjoiTGY5TFozdXRCLVNjTmY0M3lHMDJpcXZ2NmV1ZmxoYl9CV0M3alRBcXFHOCJ9"

# Run the application
ENTRYPOINT ["./baton-keycloak"]
