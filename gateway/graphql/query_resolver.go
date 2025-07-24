package graphql

import "context"

type queryResolver struct {
	server *Server
}

func (q *queryResolver) Accounts(ctx context.Context, pagination Pagination, id *string) ([]*Account, error) {
	return nil, nil
}

func (q *queryResolver) Products(ctx context.Context, pagination Pagination, query *string, id *string) ([]*Product, error) {
	return nil, nil
}
