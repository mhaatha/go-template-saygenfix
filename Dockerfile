# Stage 1: Build the application
FROM golang:1.24.0-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# The output will be a static binary named 'web' in the /app directory.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o web ./cmd/web

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/web .

# Copy templates and migrations needed by the application at runtime
COPY internal/templates ./internal/templates
COPY internal/database/migrations ./internal/database/migrations

# It's good practice to run containers as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the port the app runs on.
# The default port is 8080, but it can be overridden by the APP_PORT environment variable.
EXPOSE 8080

# Command to run the application
# The APP_PORT must be set in the environment where you run the container.
CMD ["./web"]