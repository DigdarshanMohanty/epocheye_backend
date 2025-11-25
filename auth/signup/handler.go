package signup

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/m/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	uid := uuid.New().String()

	_, err = db.Conn.Exec(
		r.Context(),
		`INSERT INTO users (uuid, email, password_hash, created_at)
     VALUES ($1, $2, $3, $4)`,
		uid, req.Email, string(hashed), time.Now(),
	)
	if err != nil {
		if pgErr, ok := err.(*pgx.PgError); ok {
			log.Printf("Postgres error code: %s, message: %s\n", pgErr.Code, pgErr.Message)
		} else {
			log.Printf("Non-Postgres error: %+v\n", err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Signup successful",
		"uid":     uid,
	})
}
