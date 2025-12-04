package login

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/m/auth"
	"example.com/m/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JSONResponse map[string]any

type LoginResponse struct {
	Message       string `json:"message"`
	UID           string `json:"uid"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	GeneratedAt   string `json:"generatedAt"`
	AccessExpires string `json:"accessExpires"`
}

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

// Handler handles user login
// @Summary User Login
// @Description Authenticates a user and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "Invalid email or password"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /login [post]
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

	var UUID uuid.UUID
	var storedHash string
	err := db.Conn.QueryRow(
		r.Context(),
		`SELECT uuid, password_hash FROM users WHERE email=$1`,
		req.Email,
	).Scan(&UUID, &storedHash)

	if err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		jsonError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// if err != nil {
	// 	jsonError(w, "Invalid user UUID stored in DB", http.StatusInternalServerError)
	// 	return
	// }

	// Correct mapping for GenerateJWT
	_, accessToken, refreshToken, genAt, accessExp, err := auth.GenerateJWT(UUID, req.Email, 0, 0)
	if err != nil {
		log.Printf("JWT generation error: %+v\n", err)
		jsonError(w, "Could not generate tokens", http.StatusInternalServerError)
		return
	}

	// Success
	jsonSuccess(w, JSONResponse{
		"message":       "Login successful",
		"uid":           UUID,
		"accessToken":   accessToken,
		"refreshToken":  refreshToken,
		"generatedAt":   genAt.Format(time.RFC3339),
		"accessExpires": accessExp.Format(time.RFC3339),
	})
}
