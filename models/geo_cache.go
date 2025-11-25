package models

import "time"

type GeoCache struct {
	ID        int       `json:"id"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Radius    int       `json:"radius"`
	Category  string    `json:"category"`
	Response  []byte    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}
