package findplaces

type Place struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Lat            float64  `json:"lat"`
	Lon            float64  `json:"lon"`
	AddressLine1   string   `json:"address_line1"`
	AddressLine2   string   `json:"address_line2"`
	Formatted      string   `json:"formatted"`
	Street         string   `json:"street"`
	City           string   `json:"city"`
	State          string   `json:"state"`
	Country        string   `json:"country"`
	Postcode       string   `json:"postcode"`
	Categories     []string `json:"categories"`
	DistanceMeters float64  `json:"distance_meters"`
}

type FindPlacesRequest struct {
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	RadiusMeters int      `json:"radius_meters"`
	Categories   []string `json:"categories"`
	Limit        int      `json:"limit"`
}

type FindPlacesResponse struct {
	Places      []Place  `json:"places"`
	GeneratedAt string   `json:"generated_at"`
}
