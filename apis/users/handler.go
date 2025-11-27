package users

import (
	"encoding/json"
	"net/http"

	"example.com/m/db"
	"example.com/m/middleware"
)

type UpdateMeRequest struct {
	Name        *string                `json:"name"`
	Phone       *string                `json:"phone"`
	AvatarURL   *string                `json:"avatar_url"`
	Preferences map[string]interface{} `json:"preferences"`
}

// --- Utility: Get UUID from email ---
func getUUID(r *http.Request) (string, error) {
	email := middleware.GetUserEmail(r)

	var uuid string
	err := db.Conn.QueryRow(
		r.Context(),
		`SELECT uuid FROM users WHERE email=$1`,
		email,
	).Scan(&uuid)

	return uuid, err
}

// --- GET /api/user/me ---
func GetMe(w http.ResponseWriter, r *http.Request) {
	uid, err := getUUID(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var (
		email, phone, name, avatar string
		preferences                []byte
		createdAt, updatedAt       string
		lastLogin                  *string
	)

	err = db.Conn.QueryRow(
		r.Context(),
		`SELECT email, phone, name, avatar_url, preferences,
		        created_at, updated_at, last_login
		   FROM users
		  WHERE uuid=$1`,
		uid,
	).Scan(
		&email, &phone, &name, &avatar, &preferences,
		&createdAt, &updatedAt, &lastLogin,
	)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"uuid":        uid,
		"email":       email,
		"phone":       phone,
		"name":        name,
		"avatar_url":  avatar,
		"preferences": json.RawMessage(preferences),
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"last_login":  lastLogin,
	})
}

// --- PUT /api/user/me ---
func UpdateMe(w http.ResponseWriter, r *http.Request) {
	uid, err := getUUID(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Conn.Exec(
		r.Context(),
		`UPDATE users
		    SET name        = COALESCE($1, name),
		        phone       = COALESCE($2, phone),
		        avatar_url  = COALESCE($3, avatar_url),
		        preferences = COALESCE($4, preferences),
		        updated_at  = NOW()
		  WHERE uuid=$5`,
		body["name"],
		body["phone"],
		body["avatar_url"],
		body["preferences"],
		uid,
	)

	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated",
	})
}

// --- GET /api/user/badges ---
func GetBadges(w http.ResponseWriter, r *http.Request) {
	uid, err := getUUID(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows, err := db.Conn.Query(
		r.Context(),
		`SELECT badge_id, earned_at
		   FROM user_badges
		  WHERE user_uuid=$1`,
		uid,
	)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := []map[string]interface{}{}

	for rows.Next() {
		var (
			badgeID  string
			earnedAt string
		)

		rows.Scan(&badgeID, &earnedAt)

		result = append(result, map[string]interface{}{
			"badge_id":  badgeID,
			"earned_at": earnedAt,
		})
	}

	json.NewEncoder(w).Encode(result)
}

// --- GET /api/user/challenges ---
func GetChallenges(w http.ResponseWriter, r *http.Request) {
	uid, err := getUUID(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows, err := db.Conn.Query(
		r.Context(),
		`SELECT challenge_id, status, progress
		   FROM user_challenges
		  WHERE user_uuid=$1`,
		uid,
	)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := []map[string]interface{}{}

	for rows.Next() {
		var (
			cid      string
			status   string
			progress int
		)

		rows.Scan(&cid, &status, &progress)

		result = append(result, map[string]interface{}{
			"challenge_id": cid,
			"status":       status,
			"progress":     progress,
		})
	}

	json.NewEncoder(w).Encode(result)
}
