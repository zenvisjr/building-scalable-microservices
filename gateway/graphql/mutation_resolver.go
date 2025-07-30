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

	// user, ok := GetUserFromContext(ctx)
	// if !ok {
	// 	Logs.Error(ctx, "No user in incoming ctx before wrapping")
	// 	return nil, errors.New("unauthenticated: please login to create a product")
	// }
	// Logs.Info(ctx, "User "+user.Email+" is creating a product.")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// user, err := RequireAdmin(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// Logs.Info(ctx, "Admin "+user.Email+" is creating a product.")

	product, err := m.server.catalogClient.PostProduct(ctx, input.Name, input.Description, input.Price, input.Stock)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.PostProduct: "+err.Error())
		return nil, err
	}
	newProduct := &Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int(product.Stock),
	}

	return newProduct, nil
}

func (m *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// user, ok := GetUserFromContext(ctx)
	// if !ok {
	// 	return nil, errors.New("unauthenticated: please login to place an order")
	// }

	// if user.ID != input.AccountID {
	// 	Logs.Error(ctx, "Unauthorized: you can only place order for your account")
	// 	return nil, errors.New("unauthorized: you can only place order for your account")
	// }

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
	// Logs.Info(ctx, "User "+user.Email+" is creating an order.")
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
			Stock:       int(p.Stock),
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

func (m *mutationResolver) Logout(ctx context.Context, input *LogoutInput) (*LogoutResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Step 1: Get claims from context
	user, ok := GetUserFromContext(ctx)
	if !ok {
		Logs.Error(ctx, "Failed to get user claims: unauthorized")
		return nil, errors.New("unauthorized")
	}

	userId := user.ID
	role := user.Role

	Logs.Info(ctx, "User "+userId+" with role "+role+" is logging out.")

	// Step 2: Handle specific user logout (when input provided)
	if input != nil && input.UserID != "" {
		targetUserId := input.UserID
		
		// Check if user is trying to logout someone else
		if userId != targetUserId {
			// Only admin can logout other users
			if role != "admin" {
				Logs.Warn(ctx, "Unauthorized logout attempt by user: "+userId)
				return &LogoutResponse{
					Message: "unauthorized: only admin can logout other users",
				}, nil
			}
			Logs.Info(ctx, "Admin "+userId+" initiated logout for user: "+targetUserId)
		} else {
			Logs.Info(ctx, "User "+userId+" initiated self logout")
		}

		// Logout the target user
		_, err := m.server.AuthClient.Logout(ctx, targetUserId)
		if err != nil {
			Logs.Error(ctx, "Error from AuthClient.Logout: "+err.Error())
			return nil, err
		}

		return &LogoutResponse{
			Message: "logout successful for user: " + targetUserId,
		}, nil
	}

	// Step 3: Handle global logout (when no input provided)
	if role != "admin" {
		Logs.Error(ctx, "Unauthorized global logout attempt by user: "+userId)
		return nil, errors.New("unauthorized: only admin can logout all users from system")
	}

	Logs.Info(ctx, "Admin "+userId+" initiated global logout")
	_, err := m.server.AuthClient.Logout(ctx, "") // Empty string for global logout
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.Logout: "+err.Error())
		return nil, err
	}

	return &LogoutResponse{
		Message: "All users logged out successfully",
	}, nil
}

func (m *mutationResolver) ResetPassword(ctx context.Context, input ResetPasswordInput) (*ResetPasswordResponse, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, ok := GetUserFromContext(ctx)
	if !ok {
		Logs.Error(ctx, "Failed to get user claims: unauthorized")
		return nil, errors.New("unauthorized")
	}
	if user.Email != input.Email {
		Logs.Error(ctx, "Unauthorized reset password attempt by user: "+user.Email)
		return nil, errors.New("unauthorized: only user can reset their own password")
	}
	Logs.Info(ctx, "User "+input.Email+" is resetting their password.")
	_, err := m.server.AuthClient.ResetPassword(ctx, input.Email, input.Password, user.ID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.ResetPassword: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Password reset successful for user: " + input.Email)
	return &ResetPasswordResponse{
		Message: "Password reset successful for user: " + input.Email,
	}, nil
}

func (m *mutationResolver) DeleteProduct(ctx context.Context, id string) (bool, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.server.catalogClient.DeleteProduct(ctx, id)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.DeleteProduct: "+err.Error())
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) RestockProduct(ctx context.Context, id string, newStock int) (bool, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.server.catalogClient.RestockProduct(ctx, id, newStock)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.RestockProduct: "+err.Error())
		return false, err
	}
	return true, nil
}