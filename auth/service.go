package auth

import (
	"context"
	"errors"
	"log"
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
}
type authService struct {
	jwtManager *JWTManager
	repository Repository
}

func NewAuthService(jwtManager *JWTManager, repository Repository) Service {
	Logs := logger.GetGlobalLogger()

	Logs.LocalOnlyInfo("AuthService initialized")
	return &authService{
		jwtManager: jwtManager,
		repository: repository,
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

	// Step 1: Call AccountService to validate user
	account, err := ac.GetEmailForAuth(ctx, email)
	if err != nil {
		Logs.Error(ctx, "Account validation failed: "+err.Error())
		return nil, errors.New("failed to validate account")
	}

	// Step 2: Validate password
	if account.PasswordHash == "" || !s.jwtManager.ValidatePassword(password, account.PasswordHash) {
		Logs.Error(ctx, "Invalid password")
		return nil, errors.New("invalid password")
	}

	// Step 3: Generate Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role, account.TokenVersion)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 4: Generate Refresh Token
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

	Logs.Info(ctx, "Login successful for user: "+email)
	Logs.LocalOnlyInfo("Login successful for user: " + email)

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
		Logs.Error(ctx, "Failed to fetch account for token verification: "+err.Error())
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
	Logs.LocalOnlyInfo("Logout called for user: " + userId)

	// Step 1: Delete refresh token
	if err := s.repository.DeleteRefreshToken(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to delete refresh token: "+err.Error())
		return err
	}

	// Step 2: Increment token version in account service
	if err := ac.IncrementTokenVersion(ctx, userId); err != nil {
		Logs.Error(ctx, "Failed to increment token version: "+err.Error())
		return err
	}
	Logs.Info(ctx, "User logged out successfully: "+userId)
	Logs.LocalOnlyInfo("User logged out successfully: " + userId)
	return nil
}
