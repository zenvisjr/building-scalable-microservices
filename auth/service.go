package auth

import (
	"context"
	"errors"

	"github.com/zenvisjr/building-scalable-microservices/account" // ← gRPC client for Account
	"github.com/zenvisjr/building-scalable-microservices/auth/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/golang-jwt/jwt/v4"
)

type Service interface {
	Signup(ctx context.Context, name string, email string, password string, role string, ac *account.Client) (*pb.AuthResponse, error)
	Login(ctx context.Context, email string, password string, ac *account.Client) (*pb.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*pb.AuthResponse, error)
	VerifyToken(ctx context.Context, token string) (*UserClaims, error)
}
type authService struct {
	jwtManager *JWTManager
}

func NewAuthService(jwtManager *JWTManager) Service {
	Logs := logger.GetGlobalLogger()

	Logs.LocalOnlyInfo("AuthService initialized")
	return &authService{
		jwtManager: jwtManager,
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
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 3: Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Email, account.Role)
	if err != nil {
		Logs.Error(ctx, "Failed to generate refresh token: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Signup successful for user: "+account.Email)

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
	accessToken, err := s.jwtManager.GenerateAccessToken(account.ID, account.Email, account.Role)
	if err != nil {
		Logs.Error(ctx, "Failed to generate access token: "+err.Error())
		return nil, err
	}

	// Step 4: Generate Refresh Token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(account.ID, account.Email, account.Role)
	if err != nil {
		Logs.Error(ctx, "Failed to generate refresh token: "+err.Error())
		return nil, err
	}

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

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*pb.AuthResponse, error) {

	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("RefreshToken called")

	// Step 1: Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
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

	// Step 2: Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		Logs.Error(ctx, "Failed to parse token claims")
		return nil, errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)

	if !ok || !ok2 || !ok3 {
		Logs.Error(ctx, "Invalid claim format")
		return nil, errors.New("invalid token claims")
	}

	// Step 3: Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(sub, email, role)
	if err != nil {
		Logs.Error(ctx, "Failed to generate new access token: "+err.Error())
		return nil, err
	}

	// Step 4: Return tokens
	return &pb.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       sub,
		Email:        email,
		Role:         role,
	}, nil
}

func (s *authService) VerifyToken(ctx context.Context, token string) (*UserClaims, error) {
	Logs := logger.GetGlobalLogger()
	claims, err := s.jwtManager.VerifyAccessToken(token)
	if err != nil {
		Logs.Error(ctx, "Failed to verify token in service: "+err.Error())
		return nil, err
	}

	sub, ok := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)

	Logs.Info(ctx, "User verified in service: "+sub+" | email = "+email+" | role = "+role)
	Logs.LocalOnlyInfo("User verified in service: "+sub+" | email = "+email+" | role = "+role)

	if !ok || !ok2 || !ok3 {
		return nil, errors.New("invalid token claims")
	}

	return &UserClaims{
		ID:    sub,
		Email: email,
		Role:  role,
	}, nil
}
