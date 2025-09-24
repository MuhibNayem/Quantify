# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-backend cmd/server/main.go

# Stage 2: Create the production image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /inventory-backend .

# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["./inventory-backend"]
