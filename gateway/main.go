package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/gateway/graphql"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	var config AppConfig
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	server, err := graphql.NewGraphQLServer(config.AccountURL, config.CatalogURL, config.OrderURL)
	if err != nil {
		log.Fatal("Failed to create GraphQL server", err)
	}

	es := server.ToExecutableSchema()
	if es == nil {
		log.Fatal("Failed to create executable schema")
	}

	h := handler.New(es)
	h.Use(extension.Introspection{}) // <-- add this line
	h.AddTransport(transport.POST{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.MultipartForm{})
	// h.AddTransport(transport.WebSocket{
	// 	KeepAliveP
	// })

	http.Handle("/graphql", h)
	http.Handle("/playground", playground.Handler("zenvis", "/graphql"))

	fmt.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
