# ----------- Stage 1: Build -----------
    FROM golang:1.20-alpine AS build

    # Install required build tools
    RUN apk --no-cache add gcc g++ make git ca-certificates
    
    # Set working directory
    WORKDIR /app
    
    # Copy go.mod and go.sum first for dependency resolution
    COPY go.mod go.sum ./
    COPY vendor/ vendor/
    
    # Copy the required source code
    COPY account/ account/
    COPY catalog/ catalog/
    COPY order/ order/
    COPY gateway/ gateway/
    
# Build the gateway service (entrypoint at gateway/main.go)
    RUN go build -mod=vendor -o graph ./gateway
    
    
    # ----------- Stage 2: Run -----------
    FROM alpine:3.20
    
    # Set working directory
    WORKDIR /app
    
    # Copy the compiled binary
    COPY --from=build /graph .
    
    # Expose the GraphQL server port
    EXPOSE 8080
    
    # Run the server
    CMD ["./graph"]
    