package graphql

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/order"
)

var errInvalidParameter = errors.New("quantity must be > 0")

type mutationResolver struct {
	server *Server
}

func (m *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	account, err := m.server.accountClient.PostAccount(ctx, input.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	newAccount := &Account{
		ID:   account.ID,
		Name: account.Name,
	}
	return newAccount, nil
}

func (m *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := m.server.catalogClient.PostProduct(ctx, input.Name, input.Description, input.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newProduct := &Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}

	return newProduct, nil
}

func (m *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var products []order.OrderedProduct
	for _, p := range input.Products {
		if p.Quantity <= 0 {
			return nil, errInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ProductID: p.ID,
			Quantity: uint32(p.Quantity),
		})
	}

	order, err := m.server.orderClient.PostOrder(ctx, input.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	newOrder := &Order{
		ID: order.ID,
		CreatedAt: order.CreatedAt.Format(time.RFC1123),
	}
	
	return newOrder, nil
}
