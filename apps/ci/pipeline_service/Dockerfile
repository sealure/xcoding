# Multi-stage build for CI pipeline_service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy monorepo source
COPY . .

# Build CI pipeline service binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ci-pipeline-service ./apps/ci/pipeline_service/cmd

# Runtime image
FROM alpine:latest

# RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/ci-pipeline-service .

RUN chown -R appuser:appgroup /app
USER appuser

# Expose HTTP and gRPC ports (match defaults used by other services)
EXPOSE 50055 10055

ENTRYPOINT ["./ci-pipeline-service"]