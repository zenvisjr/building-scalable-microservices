# Stage 1: Build
FROM golang:1.21-alpine AS build

# Install required build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory inside container
WORKDIR /build

# Copy only necessary files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY account/ account/

# Build the binary from the account microservice
RUN go build -mod=vendor -o /account ./account/cmd/account

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Copy the built binary only
COPY --from=build /build/account .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./account"]

