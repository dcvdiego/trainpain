package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type StationRepository struct {
	db *sqlx.DB
}

func NewStationRepository(db *sqlx.DB) *StationRepository {
	return &StationRepository{db: db}
}

func (r *StationRepository) SearchByName(ctx context.Context, query string, limit int) ([]domain.StationSearchResult, error) {
	if limit == 0 {
		limit = 10
	}

	var stations []domain.StationSearchResult
	searchQuery := `
		SELECT id, crs_code, station_name, station_type
		FROM stations
		WHERE station_name ILIKE $1 OR crs_code ILIKE $1
		ORDER BY
			CASE
				WHEN station_name ILIKE $2 THEN 1
				WHEN station_name ILIKE $3 THEN 2
				ELSE 3
			END,
			station_name
		LIMIT $4
	`

	searchPattern := "%" + query + "%"
	startPattern := query + "%"
	exactPattern := strings.ToUpper(query)

	err := r.db.SelectContext(ctx, &stations, searchQuery,
		searchPattern, startPattern, exactPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search stations: %w", err)
	}

	return stations, nil
}

func (r *StationRepository) GetByCRS(ctx context.Context, crsCode string) (*domain.Station, error) {
	var station domain.Station
	query := `SELECT * FROM stations WHERE crs_code = $1 LIMIT 1`

	err := r.db.GetContext(ctx, &station, query, strings.ToUpper(crsCode))
	if err != nil {
		return nil, fmt.Errorf("failed to get station by CRS: %w", err)
	}

	return &station, nil
}

func (r *StationRepository) GetByID(ctx context.Context, id int) (*domain.Station, error) {
	var station domain.Station
	query := `SELECT * FROM stations WHERE id = $1 LIMIT 1`

	err := r.db.GetContext(ctx, &station, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get station by ID: %w", err)
	}

	return &station, nil
}

func (r *StationRepository) Create(ctx context.Context, station *domain.Station) error {
	query := `
		INSERT INTO stations (crs_code, tiploc, station_name, latitude, longitude, station_type, tfl_station_id, zone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		station.CRSCode, station.TIPLOC, station.StationName,
		station.Latitude, station.Longitude, station.StationType,
		station.TfLStationID, station.Zone,
	).Scan(&station.ID, &station.CreatedAt, &station.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create station: %w", err)
	}

	return nil
}
