package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nats-io/nats.go"
	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/auth"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/order"
)

type Server struct {
	accountClient *account.Client
	catalogClient *catalog.Client
	orderClient   *order.Client
	AuthClient    *auth.Client
	nats          *nats.Conn
}

func NewGraphQLServer(accountUrl, catalogURL, orderURL, authURL string) (*Server, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("NewGraphQLServer called")
	// Connect to account service
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to account service: "+err.Error())
		return nil, err
	}
	Logs.Info(context.Background(), "Connected to account service")

	// Connect to product service
	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		Logs.Error(context.Background(), "Failed to connect to catalog service: "+err.Error())
		return nil, err
	}
	Logs.Info(context.Background(), "Connected to catalog service")

	// Connect to order service
	orderClient, err := order.NewClient(orderURL)
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		Logs.Error(context.Background(), "Failed to connect to order service: "+err.Error())
		return nil, err
	}
	Logs.Info(context.Background(), "Connected to order service")

	// Connect to auth service
	authClient, err := auth.NewClient(authURL)
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		orderClient.Close()
		Logs.Error(context.Background(), "Failed to connect to auth service: "+err.Error())
		return nil, err
	}
	Logs.Info(context.Background(), "Connected to auth service")

	// Connect to NATS
	nats, err := nats.Connect("nats://nats:4222")
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		orderClient.Close()
		Logs.Error(context.Background(), "Failed to connect to NATS: "+err.Error())
		return nil, err
	}
	Logs.Info(context.Background(), "Connected to NATS")

	return &Server{
		accountClient: accountClient,
		catalogClient: catalogClient,
		orderClient:   orderClient,
		AuthClient:    authClient,
		nats:          nats,
	}, nil
}

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{
		server: s,
	}
}

func (s *Server) Query() QueryResolver {
	return &queryResolver{
		server: s,
	}
}

func (s *Server) Account() AccountResolver {
	return &accountResolver{
		server: s,
	}
}

func (s *Server) Subscription() SubscriptionResolver {
	return &subscriptionResolver{
		server: s,
	}
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}
