# Stage 1: Build
FROM golang:1.23.3-alpine AS build

# Install required build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory inside container
WORKDIR /build

# Copy only necessary files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY account/ account/
COPY logger/ logger/
COPY auth/ auth/

# Build the binary from the account microservice
RUN go build -mod vendor -o /go/bin/app ./auth/cmd/auth

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Copy the built binary only
COPY --from=build /go/bin/app .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./app"]

