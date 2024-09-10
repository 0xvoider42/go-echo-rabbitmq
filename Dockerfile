# Stage 1: Build the Go app
FROM golang:1.23.1 AS builder

# Set environment variables
ENV GO111MODULE=on
WORKDIR /app

# Copy go.mod and go.sum files for dependencies installation
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o /go-echo

# Stage 2: Create a minimal image
FROM gcr.io/distroless/base-debian11

# Set the working directory
WORKDIR /

# Copy the binary from builder stage
COPY --from=builder /go-echo /go-echo

# Expose port (for Echo app)
EXPOSE 8080

# Command to run the app
CMD ["/go-echo"]
