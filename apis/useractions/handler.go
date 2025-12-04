package useractions

import (
	"encoding/json"
	"net/http"

	"example.com/m/utils"
	"github.com/go-chi/chi/v5"
)

type PlaceRequest struct {
	PlaceID string `json:"place_id"`
}

// SavePlaceHandler saves a place for the user
// @Summary Save Place
// @Description Save a place to the user's saved places
// @Tags useractions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PlaceRequest true "Place Details"
// @Success 200 {object} map[string]string "status: saved"
// @Failure 400 {object} map[string]string "Invalid place_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/save-place [post]
func SavePlaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req PlaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlaceID == "" {
		http.Error(w, "Invalid place_id", http.StatusBadRequest)
		return
	}

	if err := SavePlace(ctx, userUUID, req.PlaceID); err != nil {
		http.Error(w, "Failed to save place", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

// RemoveSavedPlaceHandler removes a saved place
// @Summary Remove Saved Place
// @Description Remove a place from the user's saved places
// @Tags useractions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param place_id path string true "Place ID"
// @Success 200 {object} map[string]string "status: removed"
// @Failure 400 {object} map[string]string "Missing place_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/save-place/{place_id} [delete]
func RemoveSavedPlaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	placeID := chi.URLParam(r, "place_id")

	if placeID == "" {
		http.Error(w, "Missing place_id", http.StatusBadRequest)
		return
	}

	if err := RemoveSavedPlace(r.Context(), userUUID, placeID); err != nil {
		http.Error(w, "Failed to remove saved place", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

// GetSavedPlacesHandler returns saved places
// @Summary Get Saved Places
// @Description Retrieve a list of saved places for the authenticated user
// @Tags useractions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]string "saved_places"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/saved-places [get]
func GetSavedPlacesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	places, err := GetSavedPlaces(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "Failed to fetch saved places", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string][]string{"saved_places": places})
}

// LogVisitHandler logs a visit to a place
// @Summary Log Visit
// @Description Log a visit to a specific place
// @Tags useractions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PlaceRequest true "Place Details"
// @Success 200 {object} map[string]string "status: visited"
// @Failure 400 {object} map[string]string "Invalid place_id"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/visit [post]
func LogVisitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req PlaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlaceID == "" {
		http.Error(w, "Invalid place_id", http.StatusBadRequest)
		return
	}

	if err := LogVisit(r.Context(), userUUID, req.PlaceID); err != nil {
		http.Error(w, "Failed to log visit", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "visited"})
}

// GetVisitHistoryHandler returns visit history
// @Summary Get Visit History
// @Description Retrieve the visit history for the authenticated user
// @Tags useractions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]string "visit_history"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/user/visit-history [get]
func GetVisitHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	visits, err := GetVisitHistory(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string][]string{"visit_history": visits})
}
