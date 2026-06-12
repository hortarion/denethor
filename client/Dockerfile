# Use the official Golang image as the base
FROM golang:1.26.1-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /denethor-client

# Use alpine as the final image for smaller size
FROM alpine:latest

# Install ca-certificates for HTTPS support
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /denethor-client /denethor-client
COPY --from=builder /app/public /public

# Expose the port (will be set via environment variable)
EXPOSE 8090

# Run the application
CMD ["/denethor-client"]
