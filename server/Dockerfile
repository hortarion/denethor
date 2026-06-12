# Use the official Golang image as the base
FROM golang:1.25.0-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Goose
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

# Copy ALL source code including all Go files and internal packages
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /denethor-server .

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Final stage
FROM alpine:latest

# Install PostgreSQL client
RUN apk add --no-cache postgresql-client

# Set working directory
WORKDIR /app

# Copy binary, entrypoint, Goose, and SQL migrations from builder
COPY --from=builder /denethor-server /denethor-server
COPY --from=builder /entrypoint.sh /entrypoint.sh
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Create the expected SQL directory structure and copy migrations
RUN mkdir -p /app/sql/schema
COPY --from=builder /app/internal/sql/schema /app/sql/schema

# Make sure scripts are executable
RUN chmod +x /entrypoint.sh /usr/local/bin/goose

# Expose the port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/denethor-server"]
