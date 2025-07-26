package main

import (
	"log"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

func main() {
	log.Println("Logger service starting on port 9000...")
	if err := logger.ListenGRPC(9000); err != nil {
		log.Fatalf("Logger service failed: %v", err)
	}
}
