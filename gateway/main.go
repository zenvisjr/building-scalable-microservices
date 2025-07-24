package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/gateway/graphql"
)

type AppConfig struct {
	accountURL string `envservice:"ACCOUNT_SERVICE_URL"`
	catalogURL string `envservice:"CATALOG_SERVICE_URL"`
	orderURL   string `envservice:"ORDER_SERVICE_URL"`
}

func main() {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	server, err := graphql.NewGraphQLServer(config.accountURL, config.catalogURL, config.orderURL)
	if err != nil {
		log.Fatal("Failed to create GraphQL server", err)
	}

	es := server.ToExecutableSchema()

	h := handler.New(es)

	http.Handle("/graphql", h)
	http.Handle("/playground", playground.Handler("zenvis", "/graphql"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
