# Stage 1: Build the order microservice
FROM golang:1.23.3-alpine AS build

# Install necessary build tools
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory
WORKDIR /build

# Copy only required files
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY order/ order/
COPY account/ account/
COPY catalog/ catalog/
COPY logger/ logger/

# Build the Go binary using vendor folder
RUN go build -mod vendor -o /go/bin/app ./order/cmd/order

# Stage 2: Runtime image
FROM alpine:latest

WORKDIR /app

# Copy only the built binary
COPY --from=build /go/bin/app .

EXPOSE 8080

CMD ["./app"]
