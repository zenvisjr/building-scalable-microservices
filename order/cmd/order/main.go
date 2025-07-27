package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/order"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
	MailURL     string `envconfig:"MAIL_SERVICE_URL"`
}

var (
	ctx  context.Context
	Logs *logger.Logs
	err  error
)

func main() {
	ctx = context.Background()
	var err error

	// Initialize logger for order
	Logs, err = logger.InitLogger("order")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()

	Logs.Info(ctx, "Starting order service...")

	exposePrometheusMetrics(9003)
	Logs.LocalOnlyInfo("Prometheus metrics in order service listening on port 9003")

	// Load configuration
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		Logs.Fatal(ctx, "Failed to load configuration: "+err.Error())
	}
	

	// Connect to Postgres with retry
	var r order.Repository
	err = retry.Do(
		func() error {
			r, err = order.NewPostgresRepository(config.DatabaseURL)
			if err != nil {
				Logs.Error(ctx, "Failed to connect to database: "+err.Error())
			} else {
				Logs.Info(ctx, "Connected to database: "+config.DatabaseURL)
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

	// Create OrderService
	Logs.Info(ctx, "Creating order service with dependencies")
	s, err := order.NewOrderService(r)
	if err != nil {
		Logs.Fatal(ctx, "Failed to create order service: "+err.Error())
	}

	// Start gRPC server
	Logs.Info(ctx, "Starting gRPC server for order service on port 8080")
	if err := order.ListenGRPC(s, config.AccountURL, config.CatalogURL, config.MailURL, 8080); err != nil {
		Logs.Fatal(ctx, "Failed to start gRPC server: "+err.Error())
	}
}

func exposePrometheusMetrics(port int) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}()
}
