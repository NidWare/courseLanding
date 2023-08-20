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

# Copy everything from the build stage
COPY --from=build /app .

# Copy certificates (make sure they are available on your host system)
COPY /etc/letsencrypt/live/lsukhinin.site/fullchain.pem /app/cert.pem
COPY /etc/letsencrypt/live/lsukhinin.site/privkey.pem /app/key.pem

# Run the application
ENTRYPOINT ["/app/dist/main"]
