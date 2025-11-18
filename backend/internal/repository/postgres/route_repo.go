package postgres

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type RouteRepository struct {
	db *sqlx.DB
}

func NewRouteRepository(db *sqlx.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

func (r *RouteRepository) GetOrCreate(ctx context.Context, originID, destinationID int) (*domain.Route, error) {
	// Try to get existing route
	var route domain.Route
	query := `
		SELECT * FROM routes
		WHERE origin_station_id = $1 AND destination_station_id = $2
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &route, query, originID, destinationID)
	if err == nil {
		return &route, nil
	}

	// If error is not "no rows found", return the error
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query route: %w", err)
	}

	// Create new route if not found - use ON CONFLICT to handle race conditions
	routeHash := generateRouteHash(originID, destinationID)

	insertQuery := `
		INSERT INTO routes (origin_station_id, destination_station_id, route_hash, operators, is_tfl_route)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (origin_station_id, destination_station_id)
		DO UPDATE SET origin_station_id = EXCLUDED.origin_station_id
		RETURNING id, origin_station_id, destination_station_id, route_hash, operators, is_tfl_route, created_at
	`

	err = r.db.QueryRowContext(ctx, insertQuery,
		originID, destinationID, routeHash, []string{}, false,
	).Scan(&route.ID, &route.OriginStationID, &route.DestinationStationID,
		&route.RouteHash, &route.Operators, &route.IsTfLRoute, &route.CreatedAt)

	if err != nil {
		// If still fails, try one more SELECT (race condition edge case)
		selectErr := r.db.GetContext(ctx, &route, query, originID, destinationID)
		if selectErr == nil {
			return &route, nil
		}
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	return &route, nil
}

func (r *RouteRepository) GetByID(ctx context.Context, id int) (*domain.Route, error) {
	var route domain.Route
	query := `SELECT * FROM routes WHERE id = $1 LIMIT 1`

	err := r.db.GetContext(ctx, &route, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get route by ID: %w", err)
	}

	return &route, nil
}

func generateRouteHash(originID, destinationID int) string {
	data := fmt.Sprintf("%d-%d", originID, destinationID)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}
