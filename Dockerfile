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

# Set environment variables
ENV BATON_API_URL="https://auth.dev.wcs.api.weaviate.io/auth"
ENV BATON_REALM="master"
ENV BATON_KEYCLOAK_CLIENT_ID="conductor-one-spiros"
ENV BATON_KEYCLOAK_CLIENT_SECRET="rTO2fzOydikCwZu8bdbYScoWlt4urPbZ"
#ENV BATON_CLIENT_ID="aggressive-gorgon-76791@weaviate.conductor.one/ccc"
#ENV BATON_CLIENT_SECRET="secret-token:conductorone.com:v1:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6IldBTmFWaHphLXRLZ3RjV2NRTzhsRkdFbms4RUFVbWhRLTZmUXJBWFRUbkEiLCJkIjoiTGY5TFozdXRCLVNjTmY0M3lHMDJpcXZ2NmV1ZmxoYl9CV0M3alRBcXFHOCJ9"

# Run the application
ENTRYPOINT ["/app/baton-keycloak"]
