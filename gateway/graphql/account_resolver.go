package graphql

import "context"

type accountResolver struct {
	server *Server
}

func (a *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	return nil, nil
}
