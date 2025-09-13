# Build stage
FROM golang:1.21-alpine AS builder

# Install git (needed for go modules with private repos)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o otter .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and git for cloning
RUN apk --no-cache add ca-certificates git openssh-client

# Create non-root user
RUN adduser -D -g '' otter

# Set working directory
WORKDIR /workspace

# Copy the binary from builder stage
COPY --from=builder /app/otter /usr/local/bin/otter

# Change ownership
RUN chown otter:otter /workspace

# Switch to non-root user
USER otter

# Set entrypoint
ENTRYPOINT ["otter"]
CMD ["--help"]
