package users

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/m/db"
	"example.com/m/middleware"
	"example.com/m/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserProfile struct {
	UUID        string                 `json:"uuid"`
	Email       string                 `json:"email"`
	Phone       *string                `json:"phone,omitempty"`
	Name        string                 `json:"name"`
	AvatarURL   *string                `json:"avatar_url,omitempty"`
	Preferences map[string]interface{} `json:"preferences"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastLogin   *time.Time             `json:"last_login,omitempty"`
}

// GetProfileHandler returns the current user's profile
// @Summary Get User Profile
// @Description Retrieve the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfile
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/profile [get]
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uuidStr, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var profile UserProfile
	var preferences []byte

	err = db.Conn.QueryRow(ctx,
		`SELECT uuid, email, phone, name, avatar_url, preferences, created_at, updated_at, last_login
		 FROM users WHERE uuid=$1`,
		uuidStr,
	).Scan(
		&profile.UUID,
		&profile.Email,
		&profile.Phone,
		&profile.Name,
		&profile.AvatarURL,
		&preferences,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.LastLogin,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("❌ No user found for UUID: %s", uuidStr)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("🔥 DB Error fetching user profile: %+v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Decode JSONB preferences properly
	profile.Preferences = map[string]interface{}{}
	if len(preferences) > 0 {
		if err := json.Unmarshal(preferences, &profile.Preferences); err != nil {
			log.Printf("Failed to unmarshal preferences: %v", err)
			profile.Preferences = map[string]interface{}{}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

type UpdateProfileRequest struct {
	Name        *string                `json:"name"`
	Phone       *string                `json:"phone"`
	Preferences map[string]interface{} `json:"preferences"`
}

// UpdateProfileHandler updates the current user's profile
// @Summary Update User Profile
// @Description Update the authenticated user's profile details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile Updates"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/profile [put]
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuidVal := ctx.Value(middleware.UserUUIDKey)

	uuidStr, ok := uuidVal.(string)
	if !ok || uuidStr == "" {
		http.Error(w, "User UUID missing in request context", http.StatusUnauthorized)
		return
	}

	var payload UpdateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	prefsJSON, err := json.Marshal(payload.Preferences)
	if err != nil {
		http.Error(w, "Failed to serialize preferences", http.StatusInternalServerError)
		return
	}

	_, err = db.Conn.Exec(r.Context(),
		`UPDATE users SET name=COALESCE($1, name),
                        phone=COALESCE($2, phone),
                        preferences=COALESCE($3, preferences),
                        updated_at=NOW()
     WHERE uuid=$4`,
		payload.Name,
		payload.Phone,
		prefsJSON,
		uuidStr,
	)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// UploadAvatarHandler uploads a new avatar for the user
// @Summary Upload Avatar
// @Description Upload a new avatar image for the authenticated user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar File"
// @Success 200 {object} map[string]string "avatar_url"
// @Failure 400 {object} map[string]string "Invalid file"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/avatar [post]
func UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uuidStr, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5MB
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "File too big", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Avatar file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("%s_%s", uuidStr, uuid.New().String())
	url, err := middleware.UploadFile(ctx, fileBytes, fileName)
	if err != nil || url == "" {
		http.Error(w, "Failed to upload avatar", http.StatusInternalServerError)
		return
	}

	_, err = db.Conn.Exec(ctx,
		`UPDATE users SET avatar_url=$1, updated_at=NOW() WHERE uuid=$2`,
		url, uuidStr,
	)
	if err != nil {
		http.Error(w, "Failed to save avatar URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar_url": url})
}

type UserStats struct {
	Badges     int `json:"badges"`
	Challenges struct {
		Total    int            `json:"total"`
		Pending  int            `json:"pending"`
		Progress map[string]int `json:"progress_by_status"`
	} `json:"challenges"`
}

// GetStatsHandler returns user statistics
// @Summary Get User Stats
// @Description Retrieve statistics for the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserStats
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/stats [get]
func GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	uuidStr, err := utils.GetUserUUIDFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var stats UserStats

	// badges
	err = db.Conn.QueryRow(r.Context(),
		`SELECT COUNT(*) FROM user_badges WHERE user_uuid=$1`, uuidStr).Scan(&stats.Badges)
	if err != nil {
		http.Error(w, "Failed to get badges", http.StatusInternalServerError)
		return
	}

	// challenges
	stats.Challenges.Progress = make(map[string]int)
	rows, err := db.Conn.Query(r.Context(),
		`SELECT status, COUNT(*) FROM user_challenges WHERE user_uuid=$1 GROUP BY status`, uuidStr)
	if err != nil {
		http.Error(w, "Failed to get challenges", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	totalChallenges := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats.Challenges.Progress[status] = count
		totalChallenges += count
	}
	stats.Challenges.Total = totalChallenges
	stats.Challenges.Pending = stats.Challenges.Progress["pending"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
