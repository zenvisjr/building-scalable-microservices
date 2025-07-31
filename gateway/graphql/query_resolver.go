package graphql

import (
	"context"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type queryResolver struct {
	server *Server
}

func (q *queryResolver) Accounts(ctx context.Context, pagination *Pagination, id *string, name *string) ([]*Account, error) {
	Logs := logger.GetGlobalLogger()

	// user, ok := GetUserFromContext(ctx)
	// if !ok {
	// 	Logs.Error(ctx, "No user in incoming ctx before wrapping")
	// }
	// Logs.Info(ctx, "User in incoming ctx before wrapping: "+user.Email)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// user, err := RequireAdmin(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// Logs.Info(ctx, "Admin "+user.Email+" is fetching accounts.")

	if id != nil {
		res, err := q.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			Logs.Error(ctx, "Error from accountClient.GetAccount: "+err.Error())
			return nil, err
		}
		return []*Account{{
			ID:    res.ID,
			Name:  res.Name,
			Email: res.Email,
			Role:  res.Role,
			IsActive: res.IsActive,
		}}, nil
	}

	if name != nil {
		res, err := q.server.accountClient.GetEmail(ctx, *name)
		if err != nil {
			Logs.Error(ctx, "Error from accountClient.GetEmail: "+err.Error())
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
		Logs.Error(ctx, "Error from accountClient.GetAccounts: "+err.Error())
		return nil, err
	}
	var accounts []*Account
	for _, account := range accountList {
		accounts = append(accounts, &Account{
			ID:    account.ID,
			Name:  account.Name,
			Email: account.Email,
			Role:  account.Role,
			IsActive: account.IsActive,
		})
	}
	return accounts, nil
}

func (q *queryResolver) Products(ctx context.Context, pagination *Pagination, query *string, id *string) ([]*Product, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// user, err := RequireAdmin(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// Logs.Info(ctx, "Admin "+user.Email+" is fetching products.")

	if id != nil {
		res, err := q.server.catalogClient.GetProduct(ctx, *id)
		if err != nil {
			Logs.Error(ctx, "Error from catalogClient.GetProduct: "+err.Error())
			return nil, err
		}
		return []*Product{{
			ID:          res.ID,
			Name:        res.Name,
			Description: res.Description,
			Price:       res.Price,
			Stock:       int(res.Stock),
			Sold:        int(res.Sold),
			OutOfStock:  res.OutOfStock,
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
		Logs.Error(ctx, "Error from catalogClient.GetProducts: "+err.Error())
		return nil, err
	}
	var products []*Product
	for _, product := range productList {
		products = append(products, &Product{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int(product.Stock),
			Sold:        int(product.Sold),
			OutOfStock:  product.OutOfStock,
		})
	}
	return products, nil

}

func (p Pagination) bounds() (uint64, uint64) {
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

func (q *queryResolver) CurrentUsers(ctx context.Context, pagination *Pagination, role *string) ([]*Account, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user, err := RequireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	Logs.Info(ctx, "Admin "+user.Email+" is fetching current users.")

	var (
		skip uint64
		take uint64
	)
	if pagination != nil {
		skip, take = pagination.bounds()
	}
	var ro string
	if role != nil {
		ro = *role
	}

	resp, err := q.server.AuthClient.GetCurrent(ctx, skip, take, ro)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.CurrentUsers: "+err.Error())
		return nil, err
	}
	var accounts []*Account

	if resp == nil || resp.Users == nil {
		return []*Account{}, nil
	}

	for _, account := range resp.Users {
		accounts = append(accounts, &Account{
			ID:    account.Id,
			Name:  account.Name,
			Email: account.Email,
			Role:  account.Role,
		})
	}
	return accounts, nil
}
