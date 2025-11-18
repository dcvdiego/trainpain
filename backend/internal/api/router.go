package api

import (
	"github.com/dcvdiego/trainpain-backend/internal/api/handlers"
	"github.com/dcvdiego/trainpain-backend/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	healthHandler *handlers.HealthHandler,
	stationHandler *handlers.StationHandler,
	routeHandler *handlers.RouteHandler,
) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", healthHandler.Health)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Stations
		v1.GET("/stations", stationHandler.SearchStations)
		v1.GET("/stations/:crs", stationHandler.GetStation)

		// Routes
		v1.POST("/routes/reliability", routeHandler.GetReliability)
	}

	return router
}
