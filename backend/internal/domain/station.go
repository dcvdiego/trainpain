package domain

import "time"

type Station struct {
	ID            int       `db:"id" json:"id"`
	CRSCode       *string   `db:"crs_code" json:"crs_code,omitempty"`
	TIPLOC        *string   `db:"tiploc" json:"tiploc,omitempty"`
	StationName   string    `db:"station_name" json:"station_name"`
	Latitude      *float64  `db:"latitude" json:"latitude,omitempty"`
	Longitude     *float64  `db:"longitude" json:"longitude,omitempty"`
	StationType   *string   `db:"station_type" json:"station_type,omitempty"`
	TfLStationID  *string   `db:"tfl_station_id" json:"tfl_station_id,omitempty"`
	Zone          *string   `db:"zone" json:"zone,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type StationSearchResult struct {
	ID          int     `json:"id"`
	CRSCode     *string `json:"crs_code,omitempty"`
	StationName string  `json:"station_name"`
	StationType *string `json:"station_type,omitempty"`
}
