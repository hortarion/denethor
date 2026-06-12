# Use the official Golang image as the base
FROM golang:1.25.0-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /denethor-server

# Use alpine as the final image for smaller size
FROM alpine:latest

# Install PostgreSQL client
RUN apk add --no-cache postgresql-client

# Copy the binary from the builder stage
COPY --from=builder /denethor-server /denethor-server

# Expose the port (will be set via environment variable)
EXPOSE 8080

# Run the application
CMD ["/denethor-server"]
