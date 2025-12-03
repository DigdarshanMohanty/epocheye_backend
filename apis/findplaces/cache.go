package findplaces

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func BuildCacheKey(lat, lon float64, radius int, category string) string {
	return fmt.Sprintf("%.4f:%.4f:%d:%s", lat, lon, radius, category)
}

func GetCachedResponse(ctx context.Context, conn *pgx.Conn, key string) (FindPlacesResponse, float64, float64, bool) {
	var raw []byte
	var lat, lon float64
	err := conn.QueryRow(ctx,
		`SELECT response, latitude, longitude FROM poi_cache WHERE cache_key=$1`,
		key,
	).Scan(&raw, &lat, &lon)

	if err != nil {
		return FindPlacesResponse{}, 0, 0, false
	}

	var res FindPlacesResponse
	_ = json.Unmarshal(raw, &res)
	return res, lat, lon, true
}

func CacheResponse(ctx context.Context, conn *pgx.Conn, key string,
	lat, lon float64, radius int, category string, res FindPlacesResponse) error {

	raw, _ := json.Marshal(res)

	_, err := conn.Exec(ctx,
		`INSERT INTO poi_cache(cache_key,latitude,longitude,radius,category,response)
		VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(cache_key) DO UPDATE
		SET response=EXCLUDED.response`,
		key, lat, lon, radius, category, raw,
	)
	return err
}
