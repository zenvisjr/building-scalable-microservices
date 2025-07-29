package main

import (
	"context"

	"github.com/zenvisjr/building-scalable-microservices/auth"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)


func main() {

	ctx := context.Background()
	var err error

	// Initialize logger for order
	Logs, err := logger.InitLogger("auth")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()
	Logs.LocalOnlyInfo("Starting Auth Service...")

	jwtManager := auth.NewJWTManager("a-string-secret-at-least-256-bits-long", "refresh-secret-key")


	// Create the core AuthService
	service := auth.NewAuthService(jwtManager)

	// Start gRPC server
	Logs.Info(ctx, "Starting gRPC server for auth service on port 8080")
	if err := auth.ListenGRPC(service, 8080); err != nil {
		Logs.Error(context.Background(), "Failed to start Auth gRPC server: "+err.Error())
		return
	}
}
