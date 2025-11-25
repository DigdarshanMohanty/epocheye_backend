package login

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/m/auth"
	"example.com/m/db"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JSONResponse map[string]interface{}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(JSONResponse{"error": message})
}

func jsonSuccess(w http.ResponseWriter, payload JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payload)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var uid, storedHash string
	err := db.Conn.QueryRow(
		r.Context(),
		`SELECT uuid, password_hash FROM users WHERE email=$1`,
		req.Email,
	).Scan(&uid, &storedHash)

	if err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate access + refresh tokens
	accessToken, refreshToken, genAt, accessExp, err := auth.GenerateJWT(req.Email, 0, 0)
	if err != nil {
		log.Printf("JWT generation error: %+v\n", err)
		jsonError(w, "Could not generate tokens", http.StatusInternalServerError)
		return
	}

	// Success response
	jsonSuccess(w, JSONResponse{
		"message":       "Login successful",
		"uid":           uid,
		"accessToken":   accessToken,
		"refreshToken":  refreshToken,
		"generatedAt":   genAt.Format(time.RFC3339),
		"accessExpires": accessExp.Format(time.RFC3339),
	})
}
