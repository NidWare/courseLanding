# Build stage
FROM golang:1.20 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o dist/main main.go

# Final stage
FROM golang:1.20

WORKDIR /app

# Assuming the certs will be mounted at /etc/letsencrypt
# No need to copy them here

# Run the application
ENTRYPOINT ["/app/dist/main"]
