package main

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/auth"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

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

	// Load configuration
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		Logs.Fatal(ctx, "Failed to load configuration: "+err.Error())
	}

	var r auth.Repository

	// Retry DB connection
	err = retry.Do(
		func() error {
			var err error
			r, err = auth.NewPostgresRepository(config.DatabaseURL)
			if err != nil {
				Logs.Warn(ctx, "Failed to connect to refresh_tokens database: "+err.Error())
			} else {
				Logs.Info(ctx, "Connected to refresh_tokens database: "+config.DatabaseURL)
			}
			return err
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
	)

	if err != nil {
		Logs.Fatal(ctx, "Unrecoverable refresh_tokens DB error: "+err.Error())
	}
	defer r.Close()

	// Create the core AuthService
	service := auth.NewAuthService(jwtManager, r)

	// Start gRPC server
	Logs.Info(ctx, "Starting gRPC server for auth service on port 8080")
	if err := auth.ListenGRPC(service, 8080); err != nil {
		Logs.Error(context.Background(), "Failed to start Auth gRPC server: "+err.Error())
		return
	}
}
