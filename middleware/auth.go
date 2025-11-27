package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/m/auth" // <-- adjust to your module path
)

type ctxKey string

const emailKey ctxKey = "jwtEmail"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate using your JWT file
		claims, err := auth.ValidateJWT(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract email from claims
		raw := claims["email"]
		email, ok := raw.(string)
		if !ok || email == "" {
			http.Error(w, "token missing email", http.StatusUnauthorized)
			return
		}

		// Add email to context
		ctx := context.WithValue(r.Context(), emailKey, email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to get the email anywhere
func GetUserEmail(r *http.Request) string {
	v := r.Context().Value(emailKey)
	if v == nil {
		return ""
	}
	return v.(string)
}
