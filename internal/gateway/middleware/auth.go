package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Sene4ka/cloud_storage/internal/api"
)

const (
	UserIDKey = "userID"
	EmailKey  = "email"
	TokenKey  = "token"
)

func WithAuth(next http.HandlerFunc, authClient api.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		resp, err := authClient.ValidateToken(r.Context(), &api.ValidateTokenRequest{Token: token})
		if err != nil || !resp.Valid {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, resp.UserId)
		ctx = context.WithValue(ctx, EmailKey, resp.Email)
		ctx = context.WithValue(ctx, TokenKey, token)

		next(w, r.WithContext(ctx))
	}
}
