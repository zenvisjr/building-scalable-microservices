package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/mail"
	"github.com/nats-io/nats.go"
)

func main() {
	ctx := context.Background()

	// Initialize the centralized logger
	Logs, err := logger.InitLogger("mail")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer Logs.Close()

	// Local trace of service start
	Logs.LocalOnlyInfo("Mail service bootstrapping...")

	// Notify centralized logger
	// Logs.Info(ctx, "Starting mail service")
	Logs.LocalOnlyInfo("Starting mail service...")

	// ✅ Step 1: Load MailConfig from env
	var cfg mail.MailConfig
	if err := envconfig.Process("", &cfg); err != nil {
		Logs.Fatal(ctx, "Failed to load mail config: "+err.Error())
		return
	}

	//create a nats server
	Logs.LocalOnlyInfo("Connecting to NATS in email microservice")
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		Logs.Fatal(ctx, "Failed to connect to NATS: "+err.Error())
		return
	}
	Logs.LocalOnlyInfo("Connected to NATS")

	// ✅ Step 2: Pass to mail.NewMailService
	s := mail.NewMailService(cfg, nc)

	if err := s.StartEmailSubscriber(); err != nil {
		Logs.Fatal(ctx, "Failed to start NATS email subscriber: "+err.Error())
	}

	if err := mail.InitMailGRPC(s, 8080); err != nil {
		Logs.Fatal(ctx, "Failed to start gRPC server: "+err.Error())
	}

}
