# Stage 1: Build
FROM golang:1.21-alpine AS build

# Install required build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory inside container
WORKDIR /app

# Copy only necessary files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY account/ account/

# Build the binary from the account microservice
RUN go build -mod=vendor -o /go/bin/account ./account/cmd/account

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

# Copy the built binary only
COPY --from=build /go/bin/account .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./account"]

