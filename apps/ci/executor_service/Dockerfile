# Multi-stage build for CI executor_service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy monorepo source (build from repo root to include go.mod)
COPY . .
# Download deps using root go.mod
RUN go mod download

# Build executor service binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ci-executor-service ./apps/ci/executor_service/cmd

# Runtime image
FROM alpine:latest

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/ci-executor-service .

RUN chown -R appuser:appgroup /app
USER appuser

# Expose HTTP and gRPC ports
EXPOSE 50056 10056

ENTRYPOINT ["./ci-executor-service"]