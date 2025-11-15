# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install protobuf tools
RUN apk add --no-cache protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate protobuf Go code
RUN protoc --go_out=. --go_opt=paths=source_relative shared/packets.proto
RUN mv shared/packets.pb.go server/pkg/packets/

# Install sqlc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate sqlc Go code
RUN cd server/internal/server/db/config && sqlc generate

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./server/cmd

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8414

# Health check
HEALTHCHECK --interval=60s --timeout=3s --start-period=5s --retries=3 CMD curl -f http://localhost:8414/healthz || exit 1

# Run the binary
CMD ["./main"]