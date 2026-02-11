// Package middleware provides HTTP middleware
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/emreylmaz/owlrelay/relay/internal/models"
	"github.com/emreylmaz/owlrelay/relay/internal/store"
)

// Context keys
type contextKey string

const (
	TokenContextKey     contextKey = "token"
	TokenHashContextKey contextKey = "tokenHash"
)

// TokenFromContext retrieves the token from context
func TokenFromContext(ctx context.Context) *models.Token {
	if token, ok := ctx.Value(TokenContextKey).(*models.Token); ok {
		return token
	}
	return nil
}

// TokenHashFromContext retrieves the token hash from context
func TokenHashFromContext(ctx context.Context) string {
	if hash, ok := ctx.Value(TokenHashContextKey).(string); ok {
		return hash
	}
	return ""
}

// Auth creates an authentication middleware
func Auth(tokenStore *store.TokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, "Missing Authorization header")
				return
			}

			// Expect: Bearer owl_xxxxx
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeAuthError(w, "Invalid Authorization header format")
				return
			}

			tokenString := parts[1]
			if !strings.HasPrefix(tokenString, "owl_") {
				writeAuthError(w, "Invalid token format")
				return
			}

			// Validate token
			token, err := tokenStore.Validate(tokenString)
			if err != nil {
				writeAuthError(w, "Token validation failed")
				return
			}
			if token == nil {
				writeAuthError(w, "Invalid or expired token")
				return
			}

			// Compute hash for hub lookup
			tokenHash := store.HashToken(tokenString)

			// Add token and hash to context
			ctx := context.WithValue(r.Context(), TokenContextKey, token)
			ctx = context.WithValue(ctx, TokenHashContextKey, tokenHash)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + message + `"}}`))
}
