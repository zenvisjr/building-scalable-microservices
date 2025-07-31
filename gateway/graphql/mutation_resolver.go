package graphql

import (
	"context"
	"errors"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/gateway/graphql/internal/validation"
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

	validatedInput := validation.ProductInput{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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

	validatedInput := validation.OrderInput{
		AccountID: input.AccountID,
		Products:  make([]*validation.OrderedProductInput, len(input.Products)),
	}
	for i, product := range input.Products {
		validatedInput.Products[i] = &validation.OrderedProductInput{
			ID:       product.ID,
			Quantity: product.Quantity,
		}
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	validatedInput := validation.AccountInput{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		Role:     *input.Role,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}
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

	validatedInput := validation.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}
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

	validatedInput := validation.RefreshTokenInput{
		UserID: input.UserID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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

	validatedInput := validation.LogoutInput{
		UserID: input.UserID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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

	validatedInput := validation.ResetPasswordInput{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return nil, errors.New("invalid input: " + err.Error())
	}

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
	Logs.Info(ctx, "Password reset successful for user: "+input.Email)
	return &ResetPasswordResponse{
		Message: "Password reset successful for user: " + input.Email,
	}, nil
}

func (m *mutationResolver) DeleteProduct(ctx context.Context, input ProductIDInput) (bool, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.ProductIDInput{
		ProductID: input.ProductID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return false, errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.server.catalogClient.DeleteProduct(ctx, input.ProductID)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.DeleteProduct: "+err.Error())
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) RestockProduct(ctx context.Context, input RestockProductInput) (bool, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.RestockProductInput{
		ProductID: input.ProductID,
		NewStock:  input.NewStock,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return false, errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := m.server.catalogClient.RestockProduct(ctx, input.ProductID, input.NewStock)
	if err != nil {
		Logs.Error(ctx, "Error from catalogClient.RestockProduct: "+err.Error())
		return false, err
	}
	return true, nil
}

func (m *mutationResolver) DeactivateAccount(ctx context.Context, input UserIDInput) (string, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.UserIDInput{
		UserID: input.UserID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return "", errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// user, ok := GetUserFromContext(ctx)
	// if !ok {
	// 	return "", errors.New("unauthenticated: please login to deactivate account")
	// }

	// if user.ID != id {
	// 	Logs.Error(ctx, "Unauthorized: you can only deactivate your account")
	// 	return "", errors.New("unauthorized: you can only deactivate your account")
	// }

	resp, err := m.server.AuthClient.DeactivateAccount(ctx, input.UserID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.DeactivateAccount: "+err.Error())
		return "", err
	}
	return resp.Message, nil
}

func (m *mutationResolver) ReactivateAccount(ctx context.Context, input UserIDInput) (string, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.UserIDInput{
		UserID: input.UserID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return "", errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := m.server.AuthClient.ReactivateAccount(ctx, input.UserID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.ReactivateAccount: "+err.Error())
		return "", err
	}
	return resp.Message, nil
}

func (m *mutationResolver) DeleteAccount(ctx context.Context, input UserIDInput) (string, error) {
	Logs := logger.GetGlobalLogger()

	validatedInput := validation.UserIDInput{
		UserID: input.UserID,
	}

	if err := validation.ValidateStruct(validatedInput); err != nil {
		Logs.Error(ctx, "Validation failed: "+err.Error())
		return "", errors.New("invalid input: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, ok := GetUserFromContext(ctx)
	if !ok {
		return "", errors.New("unauthenticated: please login to delete account")
	}

	if user.ID != input.UserID {
		Logs.Error(ctx, "Unauthorized: you can only delete your account")
		return "", errors.New("unauthorized: you can only delete your account")
	}

	resp, err := m.server.AuthClient.DeleteAccount(ctx, input.UserID)
	if err != nil {
		Logs.Error(ctx, "Error from AuthClient.DeleteAccount: "+err.Error())
		return "", err
	}
	return resp.Message, nil
}
