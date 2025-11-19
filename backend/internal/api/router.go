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
	pineappleHandler *handlers.PineappleHandler,
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

		// Pineapple Dance Studios classes
		v1.GET("/pineapple/classes", pineappleHandler.GetAllClasses)
		v1.GET("/pineapple/classes/day/:day", pineappleHandler.GetClassesByDay)
		v1.POST("/pineapple/subscribe", pineappleHandler.Subscribe)
		v1.DELETE("/pineapple/subscribe/:id", pineappleHandler.Unsubscribe)
		v1.GET("/pineapple/subscriptions", pineappleHandler.GetUserSubscriptions)
		v1.POST("/pineapple/scrape", pineappleHandler.ScrapeClasses)
	}

	return router
}
