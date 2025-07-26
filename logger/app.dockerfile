# Stage 1: Build the logger microservice
FROM golang:1.23.3-alpine AS build

# Install build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory inside builder container
WORKDIR /build

# Copy necessary files for build
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY logger/ logger/
# Build the logger service binary
RUN go build -mod vendor -o /go/bin/app ./logger/cmd/logger

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary
COPY --from=build /go/bin/app .

# Expose the port the logger service listens on
EXPOSE 9000

# Run the logger service
CMD ["./app"]
