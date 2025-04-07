FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mcp-graph ./cmd/server

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/mcp-graph /app/mcp-graph

# Copy configuration files
COPY config.yaml.example /app/config.yaml

# Expose the MCP SSE port
EXPOSE 3000

# Run the application
ENTRYPOINT ["/app/mcp-graph"]
