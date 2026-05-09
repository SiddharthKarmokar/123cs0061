# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the api and worker binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/worker ./cmd/worker

# Run stage for API
FROM alpine:latest AS api

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/build/api .

EXPOSE 8080 9090

CMD ["./api"]
