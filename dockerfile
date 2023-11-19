# Use the official Go image from the Docker Hub
FROM golang:1.20-alpine

# Install SQLite and gcc
RUN apk add --no-cache sqlite gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the application
RUN go build -o main .

RUN ls -al /app

# Expose port 8080 to the outside world
EXPOSE 80 443

# Run the application
CMD ["./main"]
