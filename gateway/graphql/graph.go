package main

import (
	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
	"github.com/zenvisjr/building-scalable-microservices/order"
)

//with the help of server we will be able to all account, catalog, order
// type Server struct {
// 	accountClient *account.Client
// 	catalogClient *catalog.Client
// 	orderClient *order.Client
// }

func NewGraphQLServer(accountURL, catalogURL, orderURL string) (*Server, error) {
	accountCLient, err := account.NewClient(accountURL)
	if err != nil {
		return nil, err
	}
	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		return nil, err
	}
	orderClient, err := order.NewClient(orderURL)
	if err != nil {
		return nil, err
	}

	return &Server {
		accountCLient,
		catalogClient,
		orderClient,
	}, nil
}