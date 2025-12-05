# Build stage
FROM golang:1.25-alpine AS builder

# Install dependencies
# RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# Build from monorepo root using service path
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o artifact-service ./apps/artifact/cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
# RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/artifact-service .

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 10054 50054

# Run the binary
CMD ["./artifact-service"]