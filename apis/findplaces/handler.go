package findplaces

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/m/db"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Only POST allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req FindPlacesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Latitude == 0 || req.Longitude == 0 {
		http.Error(w, `{"error":"Latitude and Longitude required"}`, http.StatusBadRequest)
		return
	}
	if req.RadiusMeters <= 0 {
		req.RadiusMeters = 5000 // default radius = 5km
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	category := "tourism"
	cacheKey := BuildCacheKey(req.Latitude, req.Longitude, req.RadiusMeters, category)

	// ---- CACHE CHECK ----
	if cachedRes, cachedLat, cachedLon, ok := GetCachedResponse(r.Context(), db.Conn, cacheKey); ok {
		// Reuse if user is close to cached coordinates (~100 meters)
		if Haversine(cachedLat, cachedLon, req.Latitude, req.Longitude) < 100 {
			json.NewEncoder(w).Encode(cachedRes)
			return
		}
	}

	// ---- FETCH FROM GEOAPIFY ----
	places, err := FetchPlaces(req.Latitude, req.Longitude, req.RadiusMeters, req.Limit, category)
	if err != nil {
		log.Printf("Geoapify fetch failed: %v", err)
		http.Error(w, `{"error":"Failed fetching places"}`, http.StatusBadGateway)
		return
	}

	response := FindPlacesResponse{
		Places:      places,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// ---- SAVE TO CACHE (async best-effort) ----
	if err := CacheResponse(
		r.Context(),
		db.Conn,
		cacheKey,
		req.Latitude,
		req.Longitude,
		req.RadiusMeters,
		category,
		response,
	); err != nil {
		log.Printf("Cache save failed: %v", err)
	}

	json.NewEncoder(w).Encode(response)
}
