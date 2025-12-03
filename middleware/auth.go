package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/m/auth"
)

type ctxKey string

const (
	emailKey    ctxKey = "jwtEmail"
	UserUUIDKey ctxKey = "userUUID"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT
		claims, err := auth.ValidateJWT(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract email
		email, ok := claims["email"].(string)
		if !ok || email == "" {
			http.Error(w, "token missing email", http.StatusUnauthorized)
			return
		}

		// Extract UUID properly as string
		rawUUID, ok := claims["user_uuid"]
		if !ok {
			http.Error(w, "token missing uuid", http.StatusUnauthorized)
			return
		}

		userUUID, ok := rawUUID.(string)
		if !ok || userUUID == "" {
			http.Error(w, "invalid uuid in token", http.StatusUnauthorized)
			return
		}

		// Set into context
		ctx := context.WithValue(r.Context(), UserUUIDKey, userUUID)
		ctx = context.WithValue(ctx, emailKey, email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
