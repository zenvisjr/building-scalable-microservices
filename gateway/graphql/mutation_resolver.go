package graphql

import "context"

type mutationResolver struct {
	server *Server
}

func (m *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	return nil, nil
}

func (m *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	return nil, nil
}

func (m *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	return nil, nil
}
