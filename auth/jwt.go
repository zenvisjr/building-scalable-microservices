package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
	accessSecret  string
	refreshSecret string
}

func NewJWTManager(accessSecret, refreshSecret string) *JWTManager {
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (j *JWTManager) ValidatePassword(plain, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}

func (j *JWTManager) GenerateAccessToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.accessSecret))
}

func (j *JWTManager) GenerateRefreshToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.refreshSecret))
}



// VerifyAccessToken parses and validates a JWT using accessSecret
func (j *JWTManager) VerifyAccessToken(tokenString string) (jwt.MapClaims, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Verifying access token...")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			Logs.Error(context.Background(), "Invalid token method")
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.accessSecret), nil
	})

	if err != nil {
		Logs.Error(context.Background(), "Failed to parse token: "+err.Error())
		return nil, err
	}

	// Extract claims and validate them
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		Logs.Info(context.Background(), "claim extracted user = "+claims["sub"].(string)+" | email = "+claims["email"].(string)+" | role = "+claims["role"].(string))
		Logs.LocalOnlyInfo("Access token verified successfully")
		return claims, nil
	}
	Logs.Error(context.Background(), "Invalid token")
	return nil, errors.New("invalid token")
}

