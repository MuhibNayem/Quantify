# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Create the directory structure matching the module path
WORKDIR /app/inventory/backend

# Copy go.mod and go.sum
COPY go.* ./

# Download dependencies
RUN go mod download
RUN go get github.com/lib/pq
RUN go mod tidy

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-backend ./cmd/server/main.go

# Stage 2: Create the production image
FROM alpine:latest

RUN adduser -D -g '' nonroot

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /inventory-backend /app/inventory-backend

USER nonroot

# Expose port 8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/inventory-backend"]
