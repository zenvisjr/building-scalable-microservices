# Stage 1: Build the order microservice
FROM golang:1.21-alpine as build

# Install necessary build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /app

# Copy only required files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY order/ order/

# Build the Go binary using vendor folder
RUN go build -mod=vendor -o /go/bin/order ./order/cmd/order

# Stage 2: Runtime image
FROM alpine:latest

WORKDIR /root/

# Copy only the built binary
COPY --from=build /go/bin/order .

EXPOSE 8080
CMD ["./order"]
