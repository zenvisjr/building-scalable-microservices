package main

import (
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
)

//we need to do 2 things
//1. connect to database
//2. start the gRPC server

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	var (
		r   catalog.Repository
		err error
	)

	err = retry.Do(
		func() error {

			r, err = catalog.NewElasticRepository(config.DatabaseURL)
			if err != nil {
				log.Println("Failed to connect to database", err)
			}
			return err
		},
		retry.Attempts(5),
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
	)

	if err != nil {
		log.Fatal("Unrecoverable DB error", err)
	}
	// defer r.Close()

	fmt.Println("Connected to database on ", config.DatabaseURL)
	fmt.Println("Listening on :8081...")
	s := catalog.NewCatalogService(r)
	log.Fatal(catalog.ListenGRPC(s, 8081))

}
