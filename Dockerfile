# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO (required for SQLite)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build -o wheeler .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/wheeler .

# Copy web templates and static assets
COPY --from=builder /app/internal/web/templates ./internal/web/templates
COPY --from=builder /app/internal/web/static ./internal/web/static

# Copy database schema files
COPY --from=builder /app/internal/database/schema.sql ./internal/database/schema.sql
COPY --from=builder /app/internal/database/wheel_strategy_example.sql ./internal/database/wheel_strategy_example.sql

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Expose the web server port
EXPOSE 8080

# Run the application
CMD ["./wheeler"]
