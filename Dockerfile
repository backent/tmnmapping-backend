# Build stage
FROM golang:1.21-alpine AS buildstage

# Install build dependencies
RUN apk add --no-cache git

# Install Wire CLI for dependency injection
RUN go install github.com/google/wire/cmd/wire@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate Wire dependencies
RUN wire ./injector

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tmn-backend .

# Production stage
FROM alpine:latest

# Install CA certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from build stage
COPY --from=buildstage /app/tmn-backend .
COPY --from=buildstage /app/database/migrations ./database/migrations

# Create a non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose the application port
EXPOSE 8088

# Run the application
CMD ["./tmn-backend"]
