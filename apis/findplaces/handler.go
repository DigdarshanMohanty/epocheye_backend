package findplaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"example.com/m/db"
	"github.com/jackc/pgx/v5"
)

/* -------------------------------------------------------
   MODELS
------------------------------------------------------- */

type GeoapifyResponse struct {
	Features []struct {
		Properties struct {
			Name       string   `json:"name"`
			Country    string   `json:"country"`
			State      string   `json:"state"`
			City       string   `json:"city"`
			Street     string   `json:"street"`
			Postcode   string   `json:"postcode"`
			Formatted  string   `json:"formatted"`
			Categories []string `json:"categories"`
			Distance   float64  `json:"distance"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

/* -------------------------------------------------------
   HAVERSINE: Calculate distance between two lat/lon points
------------------------------------------------------- */

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // meters

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

/* -------------------------------------------------------
   RATE LIMITER
------------------------------------------------------- */

func CanMakeRequest(ctx context.Context, conn *pgx.Conn, identifier string) bool {
	const limit = 50

	windowStart := time.Now().UTC().Truncate(time.Hour)

	var count int
	err := conn.QueryRow(ctx,
		`SELECT count FROM api_rate_limits
         WHERE identifier=$1 AND window_start=$2`,
		identifier, windowStart,
	).Scan(&count)

	if err != nil {
		_, err := conn.Exec(ctx,
			`INSERT INTO api_rate_limits(identifier, window_start, count)
			 VALUES ($1, $2, 1)`,
			identifier, windowStart,
		)
		return err == nil
	}

	if count >= limit {
		return false
	}

	_, _ = conn.Exec(ctx,
		`UPDATE api_rate_limits
         SET count = count + 1
         WHERE identifier=$1 AND window_start=$2`,
		identifier, windowStart,
	)

	return true
}

/* -------------------------------------------------------
   CACHE READ + WRITE
------------------------------------------------------- */

func GetCachedResponse(ctx context.Context, conn *pgx.Conn, cacheKey string) (FindPlacesResponse, float64, float64, bool) {
	var raw []byte
	var lat, lon float64

	err := conn.QueryRow(ctx,
		`SELECT response, latitude, longitude
         FROM poi_cache
         WHERE cache_key=$1`,
		cacheKey,
	).Scan(&raw, &lat, &lon)

	if err != nil {
		return FindPlacesResponse{}, 0, 0, false
	}

	var res FindPlacesResponse
	if e := json.Unmarshal(raw, &res); e != nil {
		return FindPlacesResponse{}, 0, 0, false
	}

	return res, lat, lon, true
}

func CacheResponse(ctx context.Context, conn *pgx.Conn, cacheKey string,
	lat, lon float64, radius int, category string, res FindPlacesResponse) error {

	raw, _ := json.Marshal(res)

	_, err := conn.Exec(ctx,
		`INSERT INTO poi_cache (cache_key, latitude, longitude, radius, category, response)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (cache_key) DO UPDATE
		 SET response = EXCLUDED.response,
		     latitude = EXCLUDED.latitude,
		     longitude = EXCLUDED.longitude`,
		cacheKey, lat, lon, radius, category, raw,
	)

	return err
}

/* -------------------------------------------------------
   MAIN HANDLER
------------------------------------------------------- */

func FindPlacesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type Request struct {
		UserID       string   `json:"user_id"`
		Latitude     float64  `json:"latitude"`
		Longitude    float64  `json:"longitude"`
		RadiusMeters int      `json:"radius_meters"`
		Categories   []string `json:"categories"`
		Limit        int      `json:"limit"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// ---- RATE LIMIT ----
	identifier := req.UserID
	if !CanMakeRequest(ctx, db.Conn, identifier) {
		http.Error(w, "Rate limit reached. Try again later.", http.StatusTooManyRequests)
		return
	}

	category := "tourism"
	if len(req.Categories) > 0 {
		category = req.Categories[0]
	}

	cacheKey := fmt.Sprintf("%.4f:%.4f:%d:%s",
		req.Latitude, req.Longitude, req.RadiusMeters, category,
	)

	// ---- CACHE CHECK ----
	cached, cachedLat, cachedLon, ok := GetCachedResponse(ctx, db.Conn, cacheKey)
	if ok {
		dist := Haversine(cachedLat, cachedLon, req.Latitude, req.Longitude)

		if dist < 100 {
			// user hasn't moved significantly → return cache
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cached)
			return
		}
	}

	// ---- FETCH FROM GEOAPIFY ----
	apiKey := os.Getenv("GEOAPIFY_API_KEY")
	if apiKey == "" {
		http.Error(w, "Missing Geoapify API Key", http.StatusInternalServerError)
		return
	}

	apiURL := fmt.Sprintf(
		"https://api.geoapify.com/v2/places?categories=%s&filter=circle:%.6f,%.6f,%d&limit=%d&apiKey=%s",
		category, req.Longitude, req.Latitude, req.RadiusMeters, req.Limit, apiKey,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Geoapify API failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var geo GeoapifyResponse
	if err := json.Unmarshal(body, &geo); err != nil {
		http.Error(w, "invalid Geoapify response", http.StatusInternalServerError)
		return
	}

	// ---- CONVERT PLACES ----
	places := make([]Place, 0)
	for i, f := range geo.Features {
		p := Place{
			ID:             fmt.Sprintf("place_%d", i+1),
			Name:           f.Properties.Name,
			Lat:            f.Geometry.Coordinates[1],
			Lon:            f.Geometry.Coordinates[0],
			Formatted:      f.Properties.Formatted,
			AddressLine1:   f.Properties.Street,
			City:           f.Properties.City,
			State:          f.Properties.State,
			Country:        f.Properties.Country,
			Postcode:       f.Properties.Postcode,
			Categories:     f.Properties.Categories,
			DistanceMeters: f.Properties.Distance,
		}
		places = append(places, p)
	}

	final := FindPlacesResponse{
		Places:      places,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// ---- UPDATE CACHE ----
	_ = CacheResponse(ctx, db.Conn, cacheKey,
		req.Latitude, req.Longitude, req.RadiusMeters, category, final)

	// ---- SEND RESPONSE ----
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(final)
}
