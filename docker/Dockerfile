# Start from a Golang base image
FROM golang:1.23.1-alpine3.20 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Copy the vendor directory
COPY vendor/ vendor/

# Copy the core function directory
COPY core/ core/

# Copy the source code
COPY *.go ./

# Build the Go app
RUN GOOS=linux go build -o taipeion_server .

# Start a new stage from scratch
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./taipeion_server"]
