package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zenvisjr/building-scalable-microservices/gateway/graphql"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

var (
	ctx  context.Context
	Logs *logger.Logs
	err  error
)

func main() {
	ctx = context.Background()

	Logs, err = logger.InitLogger("gateway")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()

	Logs.Info(ctx, "Starting GraphQL gateway service...")

	exposePrometheusMetrics(9004)
	Logs.LocalOnlyInfo("Prometheus metrics in gateway service listening on port 9004")

	// Load environment config
	var config AppConfig
	if err := envconfig.Process("", &config); err != nil {
		Logs.Fatal(ctx, "Failed to load environment config: "+err.Error())
	}

	// Create GraphQL server
	server, err := graphql.NewGraphQLServer(config.AccountURL, config.CatalogURL, config.OrderURL)
	if err != nil {
		Logs.Fatal(ctx, "Failed to create GraphQL server: "+err.Error())
	}

	es := server.ToExecutableSchema()
	if es == nil {
		Logs.Fatal(ctx, "Failed to create executable schema")
	}

	// Setup GraphQL handler
	h := handler.New(es)
	h.Use(extension.Introspection{})
	h.AddTransport(transport.POST{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.MultipartForm{})

	http.Handle("/graphql", h)
	http.Handle("/playground", playground.Handler("zenvis", "/graphql"))

	Logs.Info(ctx, "GraphQL gateway listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		Logs.Fatal(ctx, "Failed to start HTTP server: "+err.Error())
	}

}

func exposePrometheusMetrics(port int) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}()
}
