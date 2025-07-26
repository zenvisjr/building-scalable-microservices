package main

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}
var (
	ctx context.Context
	Logs *logger.Logs
	err error
)

func main() {
	ctx = context.Background()

	// Initialize the centralized logger
	Logs, err = logger.InitLogger("account"); 
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()

	// Local trace of service start
	Logs.LocalOnlyInfo("Account service bootstrapping...")

	// Notify centralized logger
	Logs.Info(ctx, "Starting account service")

	// Load configuration
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		Logs.Fatal(ctx, "Failed to load configuration: "+err.Error())
	}

	var r account.Repository

	// Retry DB connection
	err = retry.Do(
		func() error {
			var err error
			r, err = account.NewPostgresRepository(config.DatabaseURL)
			if err != nil {
				Logs.Warn(ctx, "Failed to connect to database: "+err.Error())
				Logs.LocalOnlyInfo("Retrying DB connection...")
			} else {
				Logs.LocalOnlyInfo("Connected to DB successfully.")
				Logs.Info(ctx, "Connected to database: " + config.DatabaseURL)
			}
			return err
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
	)

	if err != nil {
		Logs.Fatal(ctx, "Unrecoverable DB error: "+err.Error())
	}
	defer r.Close()

	Logs.LocalOnlyInfo("Launching gRPC server...")
	Logs.Info(ctx, "Starting gRPC server for account service on port 8080")

	s := account.NewAccountService(r)
	if err := account.ListenGRPC(s, 8080); err != nil {
		Logs.Fatal(ctx, "Failed to start gRPC server: "+err.Error())
	}
}
