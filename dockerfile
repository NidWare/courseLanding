# Use the official Golang image for both build and final image
FROM golang:1.20 AS build

# Set the working directory in the container
WORKDIR /app

# Copy go mod and sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the local package files to the container's workspace
COPY . .

# Build the application
RUN go build -o dist/main main.go

# Use the same Golang image for the final stage
FROM golang:1.20

# Set the working directory in the container
WORKDIR /app

# Copy everything from the build stage
COPY --from=build /app .

# Run the application
ENTRYPOINT ["/app/dist/main"]
