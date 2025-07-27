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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	account, err := m.server.accountClient.PostAccount(ctx, input.Name, input.Email)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	newAccount := &Account{
		ID:   account.ID,
		Name: account.Name,
		Email: account.Email,
	}
	return newAccount, nil
}

func (m *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	// ADD THIS LINE FIRST
	
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	log.Println("🚀 CreateOrder mutation called!")
	order, err := m.server.orderClient.PostOrder(ctx, input.AccountID, products)
	if err != nil {
		log.Println("❌ Error from orderClient.PostOrder:", err)
		return nil, err
	}

	newOrder := &Order{
		ID: order.ID,
		CreatedAt: order.CreatedAt.Format(time.RFC1123),
		TotalPrice: order.TotalPrice,
		Products: []*OrderedProduct{},
	}
	for _, p := range order.Products {
		newOrder.Products = append(newOrder.Products, &OrderedProduct{
			ID: p.ProductID,
			Name: p.Name,
			Description: p.Description,
			Price: p.Price,
			Quantity: int(p.Quantity),
		})
	}
	
	return newOrder, nil
}