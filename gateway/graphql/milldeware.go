package graphql

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zenvisjr/building-scalable-microservices/auth"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type contextKey string

const UserCtxKey = contextKey("user")

func AuthMiddleware(authClient *auth.Client) func(http.Handler) http.Handler {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("AuthMiddleware called")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			Logs.LocalOnlyInfo("AuthHeader: " + authHeader)
			// Case 1: No token → proceed as guest
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			// Case 2: Token present → verify
			token := strings.TrimPrefix(authHeader, "Bearer ")
			Logs.LocalOnlyInfo("Calling authClient.VerifyToken from middleware with token: " + token)
			Logs.Info(r.Context(), "Calling authClient.VerifyToken from middleware with token: "+token)
			user, err := authClient.VerifyToken(r.Context(), token)
			Logs.Info(r.Context(), "User verified in middleware: "+user.Email+" | role: "+user.Role)

			// If token is invalid → reject
			if err != nil {
				Logs.Error(r.Context(), "Token verification failed: "+err.Error())
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// ✅ Token is valid, now log and inject
			Logs.Info(r.Context(), "User authenticated: " + user.Email + " | role: " + user.Role)

			// Token is valid → inject user into context
			ctx := context.WithValue(r.Context(), UserCtxKey, user)
			r = r.WithContext(ctx)
			Logs.Info(r.Context(), "Injected user into context: " + user.Email + " | role: " + user.Role)

			next.ServeHTTP(w, r)
		})
	}
}

type UserClaims struct {
	UserID string
	Email  string
	Role   string
}

// GetUserFromContext safely extracts user info
func GetUserFromContext(ctx context.Context) (*auth.UserClaims, bool) {
	user, ok := ctx.Value(UserCtxKey).(*auth.UserClaims)
	return user, ok
}

func RequireAdmin(ctx context.Context) (*auth.UserClaims, error) {
	Logs := logger.GetGlobalLogger()
	user, ok := GetUserFromContext(ctx)
	if !ok {
		Logs.Error(ctx, "Unauthorized: Invalid token")
		return nil, fmt.Errorf("ctx is not valid")
	}
	if user == nil {
		Logs.Error(ctx, "Unauthorized: Invalid token")
		return nil, fmt.Errorf("user is not valid")
	}
	if user.Role != "admin" {
		Logs.Error(ctx, "Forbidden: Admin only")
		return nil, fmt.Errorf("forbidden: admin only")
	}
	return user, nil
}
