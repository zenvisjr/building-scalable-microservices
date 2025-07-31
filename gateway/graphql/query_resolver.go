package graphql

import (
	"context"
	"errors"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/gateway/graphql/internal/validation"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type queryResolver struct {
	server *Server
}

func (p *Pagination) bounds() (uint64, uint64) {
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

func safeBounds(p *Pagination) (uint64, uint64) {
	if p != nil {
		return p.bounds()
	}
	return 0, 100
}

func (q *queryResolver) Accounts(ctx context.Context, input *AccountsQueryInput) ([]*Account, error) {
	Logs := logger.GetGlobalLogger()
	validatedInput := validation.AccountsQueryInput{
		ID:         *input.ID,
		Name:       *input.Name,
		Pagination: &validation.Pagination{
			Skip: *input.Pagination.Skip,
			Take: *input.Pagination.Take,
		},
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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

	if input.ID != nil {
		res, err := q.server.accountClient.GetAccount(ctx, *input.ID)
		if err != nil {
			Logs.Error(ctx, "Error from accountClient.GetAccount: "+err.Error())
			return nil, err
		}
		return []*Account{{
			ID:       res.ID,
			Name:     res.Name,
			Email:    res.Email,
			Role:     res.Role,
			IsActive: res.IsActive,
		}}, nil
	}

	if input.Name != nil {
		res, err := q.server.accountClient.GetEmail(ctx, *input.Name)
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

	skip, take := safeBounds(input.Pagination)

	accountList, err := q.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		Logs.Error(ctx, "Error from accountClient.GetAccounts: "+err.Error())
		return nil, err
	}
	var accounts []*Account
	for _, account := range accountList {
		accounts = append(accounts, &Account{
			ID:       account.ID,
			Name:     account.Name,
			Email:    account.Email,
			Role:     account.Role,
			IsActive: account.IsActive,
		})
	}
	return accounts, nil
}

func (q *queryResolver) Products(ctx context.Context, input *ProductsQueryInput) ([]*Product, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.ProductsQueryInput{
		Query:      *input.Query,
		ID:         *input.ID,
		Pagination: &validation.Pagination{
			Skip: *input.Pagination.Skip,
			Take: *input.Pagination.Take,
		},
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// user, err := RequireAdmin(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// Logs.Info(ctx, "Admin "+user.Email+" is fetching products.")

	if input.ID != nil {
		res, err := q.server.catalogClient.GetProduct(ctx, *input.ID)
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

	skip, take := safeBounds(input.Pagination)

	var qu string
	if input.Query != nil {
		qu = *input.Query
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



func (q *queryResolver) CurrentUsers(ctx context.Context, input *CurrentUsersQueryInput) ([]*Account, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.CurrentUsersQueryInput{
		Role:       *input.Role,
		Pagination: &validation.Pagination{
			Skip: *input.Pagination.Skip,
			Take: *input.Pagination.Take,
		},
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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
	if input.Pagination != nil {
		skip, take = input.Pagination.bounds()
	}
	var ro string
	if input.Role != nil {
		ro = *input.Role
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

func (q *queryResolver) SuggestProducts(ctx context.Context, input *SuggestProductsQueryInput) ([]*Product, error) {
	Logs := logger.GetGlobalLogger()
	validatedInput := validation.SuggestProductsQueryInput{
		Query:  input.Query,
		Size:   *input.Size,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	s := 10
	if input.Size != nil {
		s = *input.Size
	}

	ai := false
	if input.UseAi != nil {
		ai = *input.UseAi
	}
	res, err := q.server.catalogClient.SuggestProducts(ctx, input.Query, s, ai)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.SuggestProducts: "+err.Error())
		return nil, err
	}
	var products []*Product
	for _, product := range res {
		products = append(products, &Product{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int(product.Stock),
			Sold:        int(product.Sold),
			OutOfStock:  product.OutOfStock,
			Score:       product.Score,
		})
	}
	return products, nil

}
