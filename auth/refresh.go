package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string    `json:"access_token"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse & validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	if typ, _ := claims["type"].(string); typ != "refresh" {
		http.Error(w, "invalid token type", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "invalid token data: email", http.StatusUnauthorized)
		return
	}

	userUUIDStr, ok := claims["user_uuid"].(string)
	if !ok {
		http.Error(w, "invalid token data: uuid", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userUUIDStr)
	if err != nil {
		http.Error(w, "invalid uuid format", http.StatusInternalServerError)
		return
	}

	// Generate NEW access token ONLY
	_, accessToken, _, genAt, accessExp, err :=
		GenerateJWT(userUUID, email, 0, 0)
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	resp := refreshResponse{
		AccessToken: accessToken,
		GeneratedAt: genAt,
		ExpiresAt:   accessExp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
