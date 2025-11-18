package postgres

import (
	"context"
	"crypto/md5"
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

	// Create new route if not found
	routeHash := generateRouteHash(originID, destinationID)

	insertQuery := `
		INSERT INTO routes (origin_station_id, destination_station_id, route_hash, operators, is_tfl_route)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err = r.db.QueryRowContext(ctx, insertQuery,
		originID, destinationID, routeHash, []string{}, false,
	).Scan(&route.ID, &route.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	route.OriginStationID = originID
	route.DestinationStationID = destinationID
	route.RouteHash = routeHash
	route.IsTfLRoute = false

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
