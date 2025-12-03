package useractions

import (
	"encoding/json"
	"net/http"

	"example.com/m/utils"
	"github.com/gorilla/mux"
)

type PlaceRequest struct {
	PlaceID string `json:"place_id"`
}

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

func RemoveSavedPlaceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := utils.GetUserUUIDFromCtx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	placeID := mux.Vars(r)["place_id"]

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
