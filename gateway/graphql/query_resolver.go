package graphql

import (
	"context"
	"log"
	"time"
)

type queryResolver struct {
	server *Server
}

func (q *queryResolver) Accounts(ctx context.Context, pagination *Pagination, id *string, name *string) ([]*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		res, err := q.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Account{{
			ID:   res.ID,
			Name: res.Name,
			Email: res.Email,
		}}, nil
	}

	if name != nil {
		res, err := q.server.accountClient.GetEmail(ctx, *name)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Account{
			{
				Email: res.Email,
			},
		}, nil
	}

	var (
		skip uint64
		take uint64
	)
	if pagination != nil {
		skip, take = pagination.bounds()
	}
	accountList, err := q.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var accounts []*Account
	for _, account := range accountList {
		accounts = append(accounts, &Account{
			ID:   account.ID,
			Name: account.Name,
			Email: account.Email,
		})
	}
	return accounts, nil
}

func (q *queryResolver) Products(ctx context.Context, pagination *Pagination, query *string, id *string) ([]*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if id != nil {
		res, err := q.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return []*Product{{
			ID:          res.ID,
			Name:        res.Name,
			Description: res.Description,
			Price:       res.Price,
		}}, nil
	}

	var (
		skip uint64
		take uint64
	)
	if pagination != nil {
		skip, take = pagination.bounds()
	}
	var qu string
	if query != nil {
		qu = *query
	}

	productList, err := q.server.catalogClient.GetProducts(ctx, skip, take, nil, qu)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var products []*Product
	for _, product := range productList {
		products = append(products, &Product{
			ID:   product.ID,
			Name: product.Name,
			Description: product.Description,
			Price: product.Price,
		})
	}
	return products, nil

}


func(p Pagination) bounds () (uint64, uint64) {
	skipValue := uint64(0)
	takeValue := uint64(100)
	if p.Skip != nil {
		skipValue = uint64(*p.Skip)
	} 
	if p.Take != nil {
		takeValue = uint64(*p.Take)
	}
	return skipValue, takeValue

}