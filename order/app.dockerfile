# Stage 1: Build the order microservice
FROM golang:1.21-alpine as build

# Install necessary build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /build

# Copy only required files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY order/ order/

# Build the Go binary using vendor folder
RUN go build -mod=vendor -o order ./order/cmd/order

# Stage 2: Runtime image
FROM alpine:latest

WORKDIR /app

# Copy only the built binary
COPY --from=build /build/order .

EXPOSE 8080

CMD ["./order"]
