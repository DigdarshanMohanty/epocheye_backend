package findplaces

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type geoapifyResponse struct {
	Features []struct {
		Properties struct {
			PlaceID      string   `json:"place_id"`
			Name         string   `json:"name"`
			Street       string   `json:"street"`
			City         string   `json:"city"`
			State        string   `json:"state"`
			Postcode     string   `json:"postcode"`
			Country      string   `json:"country"`
			Formatted    string   `json:"formatted"`
			AddressLine1 string   `json:"address_line1"`
			AddressLine2 string   `json:"address_line2"`
			Distance     float64  `json:"distance"`
			Categories   []string `json:"categories"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func FetchPlaces(lat, lon float64, radius, limit int, category string) ([]Place, error) {
	apiKey := os.Getenv("GEOAPIFY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing GEOAPIFY_API_KEY")
	}

	url := fmt.Sprintf(
		"https://api.geoapify.com/v2/places?categories=%s&filter=circle:%.6f,%.6f,%d&limit=%d&apiKey=%s",
		category, lon, lat, radius, limit, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var geo geoapifyResponse
	if err := json.Unmarshal(body, &geo); err != nil {
		return nil, err
	}

	places := make([]Place, 0)
	for _, f := range geo.Features {
		p := Place{
			PlaceID:        f.Properties.PlaceID,
			Name:           f.Properties.Name,
			Lat:            f.Geometry.Coordinates[1],
			Lon:            f.Geometry.Coordinates[0],
			Formatted:      f.Properties.Formatted,
			Street:         f.Properties.Street,
			AddressLine1:   f.Properties.AddressLine1,
			AddressLine2:   f.Properties.AddressLine2,
			City:           f.Properties.City,
			State:          f.Properties.State,
			Country:        f.Properties.Country,
			Postcode:       f.Properties.Postcode,
			Categories:     f.Properties.Categories,
			DistanceMeters: f.Properties.Distance,
		}

		places = append(places, p)
	}

	return places, nil
}
