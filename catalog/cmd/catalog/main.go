package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

var (
	ctx  context.Context
	Logs *logger.Logs
	err  error
)

func main() {
	ctx = context.Background()

	// Initialize logger for catalog
	Logs, err = logger.InitLogger("catalog")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()

	Logs.Info(ctx, "Starting catalog service...")

	exposePrometheusMetrics(9002)
	Logs.LocalOnlyInfo("Prometheus metrics in catalog service listening on port 9002")

	// Load env config
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		Logs.Fatal(ctx, "Failed to load configuration: "+err.Error())
	}

	// Connect to Elastic DB with retry
	var r catalog.Repository
	err = retry.Do(
		func() error {
			r, err = catalog.NewElasticRepository(config.DatabaseURL)
			if err != nil {
				Logs.Error(ctx, "Failed to connect to Elasticsearch: "+err.Error())
			} else {
				Logs.Info(ctx, "Connected to Elasticsearch at "+config.DatabaseURL)
			}
			return err
		},
		retry.Attempts(10),
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
	)

	if err != nil {
		Logs.Fatal(ctx, "Unrecoverable DB error: "+err.Error())
	}

	// 🧠 Call this to ensure index is created if not present
	if err := r.EnsureCatalogIndex(context.Background()); err != nil {
		Logs.Fatal(ctx, "Failed to ensure catalog index: "+err.Error())
	}

	// Start gRPC server
	Logs.Info(ctx, "Starting gRPC server for catalog microservice on port 8080")
	s := catalog.NewCatalogService(r)
	if err := catalog.ListenGRPC(s, 8080); err != nil {
		Logs.Fatal(ctx, "Failed to start gRPC server: "+err.Error())
	}

}

func exposePrometheusMetrics(port int) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}()
}
