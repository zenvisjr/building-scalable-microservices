package auth

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zenvisjr/building-scalable-microservices/account" // ← gRPC client for Account
	"github.com/zenvisjr/building-scalable-microservices/auth/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Service interface {
	Signup(ctx context.Context, name string, email string, password string, role string, ac *account.Client) (*pb.AuthResponse, error)
	Login(ctx context.Context, email string, password string, ac *account.Client) (*pb.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*pb.AuthResponse, error)
	VerifyToken(ctx context.Context, token string, ac *account.Client) (*UserClaims, error)

	Logout(ctx context.Context, userId string, ac *account.Client) error
	// LogoutAll(ctx context.Context, userId string) error
	GetCurrentUsers(ctx context.Context, ac *account.Client, skip uint64, take uint64, role string) ([]*User, error)
	ResetPasswordForAccount(ctx context.Context, email string, password string, userId string, ac *account.Client) (*pb.AuthResponse, error)
	DeactivateAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error)
	ReactivateAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error)
	DeleteAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error)
}

type User struct {
	Id    string
	Name  string
	Email string
	Role  string
}
type authService struct {
	jwtManager    *JWTManager
	repository    Repository
	loggedInUsers map[string]bool
	mu            sync.Mutex
}

func NewAuthService(jwtManager *JWTManager, repository Repository) Service {
	Logs := logger.GetGlobalLogger()

	Logs.LocalOnlyInfo("AuthService initialized")
	return &authService{
		jwtManager:    jwtManager,
		repository:    repository,
		loggedInUsers: make(map[string]bool),
	}
}

func (s *authService) Signup(ctx context.Context, name string, email string, password string, role string, ac *account.Client) (*pb.AuthResponse, error) {

	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Signup called for email: " + email)

	//default role is user
	if role == "" {
		role = "user"
	}
	// Step 1: Call AccountService to create user (password will be hashed there)
	account, err := ac.PostAccount(ctx, name, email, password, role)
	if err != nil {
		Logs.Error(ctx, "Account creation failed: "+err.Error())
		return nil, errors.New("failed to create account")
	}

	// Step 2: Generate Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 3: Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate refresh token: "+err.Error())
		return nil, err
	}

	// Step 5: Store Refresh Token in DB
	if err := s.repository.StoreRefreshToken(ctx, refreshToken, account.ID); err != nil {
		Logs.Error(ctx, "Failed to store refresh token: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Refresh token stored successfully for user: "+email)
	Logs.LocalOnlyInfo("Refresh token stored successfully for user: " + email)

	Logs.Info(ctx, "Signup successful for user: "+account.Email)
	Logs.LocalOnlyInfo("Signup successful for user: " + account.Email)

	//add user to loggedInUser map to track it later
	s.mu.Lock()
	s.loggedInUsers[account.ID] = true
	s.mu.Unlock()

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       account.ID,
		Email:        account.Email,
		Role:         account.Role,
	}, nil
}

func (s *authService) Login(ctx context.Context, email string, password string, ac *account.Client) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Login called for email: " + email)

	if email == "" || password == "" {
		Logs.Error(ctx, "Invalid email or password")
		return nil, errors.New("invalid email or password")
	}

	if _, ok := s.loggedInUsers[email]; ok {
		Logs.Error(ctx, "User already logged in")
		return nil, errors.New("user already logged in")
	}

	// Step 1: Call AccountService to validate user
	account, err := ac.GetEmailForAuth(ctx, email)
	if err != nil {
		Logs.Error(ctx, "Account validation failed: "+err.Error())
		return nil, errors.New("failed to validate account")
	}

	// Step 2: Check if account is active
	if !account.IsActive {
		Logs.Error(ctx, "Account is not active")
		return nil, errors.New("account is not active")
	}

	// Step 3: Validate password
	if account.PasswordHash == "" || !s.jwtManager.ValidatePassword(password, account.PasswordHash) {
		Logs.Error(ctx, "Invalid password")
		return nil, errors.New("invalid password")
	}

	// Step 4: Generate Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 5: Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate refresh token: "+err.Error())
		return nil, err
	}

	// Step 6: Store Refresh Token in DB
	if err := s.repository.StoreRefreshToken(ctx, refreshToken, account.ID); err != nil {
		Logs.Error(ctx, "Failed to store refresh token: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Refresh token stored successfully for user: "+email)
	Logs.LocalOnlyInfo("Refresh token stored successfully for user: " + email)

	Logs.Info(ctx, "Login successful for user: "+email)
	Logs.LocalOnlyInfo("Login successful for user: " + email)

	//add user to loggedInUser map to track it later
	s.mu.Lock()
	s.loggedInUsers[account.ID] = true
	s.mu.Unlock()

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       account.ID,
		Email:        account.Email,
		Role:         account.Role,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, userId string) (*pb.AuthResponse, error) {

	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("RefreshToken called")

	// Step 1: Get refresh token from DB using userId
	tokenData, err := s.repository.GetRefreshToken(ctx, userId)
	if err != nil {
		Logs.Error(ctx, "Failed to get refresh token: "+err.Error())
		return nil, errors.New("failed to get refresh token")
	}

	// Step 2: Check expiration from DB
	if time.Now().After(tokenData.ExpiresAt) {
		Logs.Error(ctx, "Refresh token has expired")
		return nil, errors.New("refresh token expired, please login again")
	}

	// Step 3: Parse and validate refresh token
	token, err := jwt.Parse(tokenData.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Make sure it's signed with the expected method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtManager.refreshSecret), nil
	})
	if err != nil || !token.Valid {
		Logs.Error(ctx, "Invalid or expired refresh token: "+err.Error())
		return nil, errors.New("invalid or expired refresh token")
	}

	// Step 4: Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		Logs.Error(ctx, "Failed to parse token claims")
		return nil, errors.New("invalid token claims")
	}

	sub, ok1 := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)

	//numbers are float32 by default
	tokenVersionFloat, ok4 := claims["token_version"].(float64)
	if !ok4 {
		Logs.Error(ctx, "Invalid claim format: token_version not a float64")
		return nil, errors.New("invalid token claims")
	}
	tokenVersion := int32(tokenVersionFloat) // convert safely

	if !ok1 || !ok2 || !ok3 || !ok4 {
		Logs.Error(ctx, "Invalid claim format")
		return nil, errors.New("invalid token claims")
	}

	// Step 5: Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(sub, email, role, tokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate new access token: "+err.Error())
		return nil, err
	}

	// Step 6: Generate new refresh token (rotation)
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(sub, email, role, tokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to rotate refresh token: "+err.Error())
		return nil, err
	}

	// Step 7: Store new refresh token in DB
	if err := s.repository.StoreRefreshToken(ctx, newRefreshToken, sub); err != nil {
		Logs.Error(ctx, "Failed to store new refresh token: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Refresh token rotated successfully for user: "+email)

	// Step 8: Return tokens
	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserId:       sub,
		Email:        email,
		Role:         role,
	}, nil
}

func (s *authService) VerifyToken(ctx context.Context, token string, ac *account.Client) (*UserClaims, error) {
	Logs := logger.GetGlobalLogger()
	// Step 1: Verify token
	claims, err := s.jwtManager.VerifyAccessToken(token)
	if err != nil {
		Logs.Error(ctx, "Failed to verify token in service: "+err.Error())
		return nil, err
	}

	sub, ok := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)
	tokenVersionFloat, ok4 := claims["token_version"].(float64)
	log.Printf("tokenVersionFloat: %f", tokenVersionFloat)

	if !ok || !ok2 || !ok3 || !ok4 {
		return nil, errors.New("invalid token claims")
	}
	tokenVersion := int32(tokenVersionFloat)

	// Step 2: Get latest account info from DB
	account, err := ac.GetEmailForAuth(ctx, email) // or use userID if you have it
	if err != nil {
		Logs.Error(ctx, "Failed to fetch account for token verification, account does not exist: "+err.Error())
		return nil, err
	}

	log.Printf("tokenVersion: %d", tokenVersion)
	log.Printf("account.TokenVersion: %d", account.TokenVersion)
	// Step 3: Compare token versions
	if tokenVersion != account.TokenVersion {
		Logs.Error(ctx, "Token version mismatch - user logged out, please login again")
		return nil, errors.New("token invalid or expired, please login again")
	}

	Logs.Info(ctx, "User verified in service: "+sub+" | email = "+email+" | role = "+role)
	Logs.LocalOnlyInfo("User verified in service: " + sub + " | email = " + email + " | role = " + role)

	return &UserClaims{
		ID:    sub,
		Email: email,
		Role:  role,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userId string, ac *account.Client) error {
	Logs := logger.GetGlobalLogger()

	if userId != "" {
		Logs.LocalOnlyInfo("Logout called for user: " + userId)

		// Step 1: Delete all refresh tokens
		if err := s.repository.DeleteRefreshToken(ctx, userId); err != nil {
			Logs.Error(ctx, "Failed to delete refresh token: "+err.Error())
			return err
		}

		// Step 2: Increment token version to invalidate tokens
		if err := ac.IncrementTokenVersion(ctx, userId); err != nil {
			Logs.Error(ctx, "Failed to increment token version: "+err.Error())
			return err
		}

		// Step 3: Remove from cache
		s.mu.Lock()
		delete(s.loggedInUsers, userId)
		s.mu.Unlock()

		Logs.Info(ctx, "User logged out successfully: "+userId)
		return nil
	}

	Logs.Info(ctx, "Global logout initiated")
	for userId := range s.loggedInUsers {
		Logs.LocalOnlyInfo("Logout called for user: " + userId)

		// Step 1: Delete all refresh tokens
		if err := s.repository.DeleteRefreshToken(ctx, userId); err != nil {
			Logs.Error(ctx, "Failed to delete refresh token: "+err.Error())
			return err
		}

		// Step 2: Increment token version to invalidate tokens
		if err := ac.IncrementTokenVersion(ctx, userId); err != nil {
			Logs.Error(ctx, "Failed to increment token version: "+err.Error())
			return err
		}

		Logs.Info(ctx, "User logged out successfully: "+userId)

	}
	// Clear the entire map after logout
	s.mu.Lock()
	s.loggedInUsers = make(map[string]bool)
	s.mu.Unlock()

	Logs.Info(ctx, "All users logged out successfully")

	return nil
}

func (s *authService) GetCurrentUsers(ctx context.Context, ac *account.Client, skip uint64, take uint64, role string) ([]*User, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("CurrentUsers called")

	var users []*User
	var count uint64 = 0

	for userId := range s.loggedInUsers {
		// Apply pagination skipping
		if count < skip {
			count++
			continue
		}
		if take > 0 && uint64(len(users)) >= take {
			break
		}

		accountData, err := ac.GetAccount(ctx, userId)
		if err != nil {
			Logs.Error(ctx, "Failed to fetch account for userID "+userId+": "+err.Error())
			continue // Don't fail everything, just skip this user
		}

		// Filter by role if specified
		if role != "" && accountData.Role != role {
			continue
		}

		users = append(users, &User{
			Id:    userId,
			Name:  accountData.Name,
			Email: accountData.Email,
			Role:  accountData.Role,
		})
		count++
	}

	return users, nil
}

func (s *authService) ResetPasswordForAccount(ctx context.Context, email string, password string, userId string, ac *account.Client) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "ResetPasswordForAccount called for email: "+email)

	// Step 1: Call AccountService to create user (password will be hashed there)
	err := ac.UpdatePassword(ctx, email, password)
	if err != nil {
		Logs.Error(ctx, "Password reset failed: "+err.Error())
		return nil, err
	}

	// Step 1: Delete all refresh tokens
	if err := s.repository.DeleteRefreshToken(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to delete refresh token: "+err.Error())
		return nil, err
	}

	// Step 2: Increment token version to invalidate tokens
	if err := ac.IncrementTokenVersion(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to increment token version: "+err.Error())
		return nil, err
	}

	// Step 2: Get latest account info from DB
	account, err := ac.GetEmailForAuth(ctx, email)
	if err != nil {
		Logs.Error(ctx, "Failed to fetch account for auth: "+err.Error())
		return nil, err
	}

	// Step 2: Generate Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 3: Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate refresh token: "+err.Error())
		return nil, err
	}

	// Step 5: Store Refresh Token in DB
	if err := s.repository.StoreRefreshToken(ctx, refreshToken, account.ID); err != nil {
		Logs.Error(ctx, "Failed to store refresh token: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Refresh token stored successfully for user: "+email)
	Logs.LocalOnlyInfo("Refresh token stored successfully for user: " + email)

	Logs.Info(ctx, "Password reset successful for user: "+account.Email)
	Logs.LocalOnlyInfo("Password reset successful for user: " + account.Email)

	//add user to loggedInUser map to track it later
	s.mu.Lock()
	s.loggedInUsers[account.ID] = true
	s.mu.Unlock()

	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       account.ID,
		Email:        account.Email,
		Role:         account.Role,
	}, nil
}

func (s *authService) DeactivateAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Deactivating account for user: "+userId)
	if err := ac.DeactivateAccount(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to deactivate account: "+err.Error())
		return nil, err
	}

	if err := s.Logout(ctx, userId, ac); err != nil {
		Logs.Error(ctx, "Failed to logout: "+err.Error())
		return nil, err
	}

	return &pb.UpdateAccountResponse{Message: "Account deactivated successfully, Please reactivate to use again"}, nil
}

func (s *authService) ReactivateAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Reactivating account for user: "+userId)
	if err := ac.ReactivateAccount(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to reactivate account: "+err.Error())
		return nil, err
	}
	return &pb.UpdateAccountResponse{Message: "Account reactivated successfully, Please login again"}, nil
}

func (s *authService) DeleteAccount(ctx context.Context, userId string, ac *account.Client) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Deleting account for user: "+userId)
	if err := ac.DeleteAccount(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to delete account: "+err.Error())
		return nil, err
	}

	// Step 1: Delete all refresh tokens
	if err := s.repository.DeleteRefreshToken(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to delete refresh token: "+err.Error())
		return nil, err
	}

	// Step 2: Remove from cache
	s.mu.Lock()
	delete(s.loggedInUsers, userId)
	s.mu.Unlock()

	return &pb.UpdateAccountResponse{Message: "Account deleted successfully"}, nil
}
