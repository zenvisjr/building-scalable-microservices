package graphql

import (
	"context"
	"errors"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/order"
)

var errInvalidParameter = errors.New("quantity must be > 0")

type mutationResolver struct {
	server *Server
}

// func (m *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()
// 	account, err := m.server.accountClient.PostAccount(ctx, input.Name, input.Email, input.Password, input.Role)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}

// 	newAccount := &Account{
// 		ID:   account.ID,
// 		Name: account.Name,
// 		Email: account.Email,
// 		Role: account.Role,
// 	}
// 	return newAccount, nil
// }

func (m *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	Logs := logger.GetGlobalLogger()

	user, ok := GetUserFromContext(ctx)
	if !ok {
		Logs.Error(ctx, "No user in incoming ctx before wrapping")
		return nil, errors.New("unauthenticated: please login to create a product")
	}
	Logs.Info(ctx, "User "+user.Email+" is creating a product.")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := RequireAdmin(ctx)
	if err != nil {
		return nil, err
	}

	Logs.Info(ctx, "Admin "+user.Email+" is creating a product.")

	product, err := m.server.catalogClient.PostProduct(ctx, input.Name, input.Description, input.Price)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.PostProduct: "+err.Error())
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
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, ok := GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthenticated: please login to place an order")
	}

	var products []order.OrderedProduct
	for _, p := range input.Products {
		if p.Quantity <= 0 {
			return nil, errInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ProductID: p.ID,
			Quantity:  uint32(p.Quantity),
		})
	}
	Logs.Info(ctx, "User "+user.Email+" is creating an order.")
	order, err := m.server.orderClient.PostOrder(ctx, input.AccountID, products)
	if err != nil {
		Logs.Error(ctx, "Error from orderClient.PostOrder: "+err.Error())
		return nil, err
	}

	newOrder := &Order{
		ID:         order.ID,
		CreatedAt:  order.CreatedAt.Format(time.RFC1123),
		TotalPrice: order.TotalPrice,
		Products:   []*OrderedProduct{},
	}
	for _, p := range order.Products {
		newOrder.Products = append(newOrder.Products, &OrderedProduct{
			ID:          p.ProductID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    int(p.Quantity),
		})
	}

	return newOrder, nil
}

func (m *mutationResolver) Signup(ctx context.Context, input AccountInput) (*AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	authResp, err := m.server.AuthClient.Signup(ctx, input.Name, input.Email, input.Password, *input.Role)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.Signup: "+err.Error())
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		UserID:       authResp.UserId,
		Email:        authResp.Email,
		Role:         authResp.Role,
	}, nil
}

func (m *mutationResolver) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	authResp, err := m.server.AuthClient.Login(ctx, input.Email, input.Password)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.Login: "+err.Error())
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		UserID:       authResp.UserId,
		Email:        authResp.Email,
		Role:         authResp.Role,
	}, nil

}

func (m *mutationResolver) RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	authResp, err := m.server.AuthClient.RefreshToken(ctx, input.UserID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.RefreshToken: "+err.Error())
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		UserID:       authResp.UserId,
		Email:        authResp.Email,
		Role:         authResp.Role,
	}, nil
}

func (m *mutationResolver) Logout(ctx context.Context, input LogoutInput) (*LogoutResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	Logs.Info(ctx, "User "+input.UserID+" is logging out.")

	authResp, err := m.server.AuthClient.Logout(ctx, input.UserID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.Logout: "+err.Error())
		return nil, err
	}

	return &LogoutResponse{
		Message: authResp.Message,
	}, nil
}
