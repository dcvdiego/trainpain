package handlers

import (
	"net/http"

	"github.com/dcvdiego/trainpain-backend/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StationHandler struct {
	stationRepo *postgres.StationRepository
	logger      *zap.Logger
}

func NewStationHandler(stationRepo *postgres.StationRepository, logger *zap.Logger) *StationHandler {
	return &StationHandler{
		stationRepo: stationRepo,
		logger:      logger,
	}
}

// SearchStations handles GET /api/v1/stations?q=search_term
func (h *StationHandler) SearchStations(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	stations, err := h.stationRepo.SearchByName(c.Request.Context(), query, 20)
	if err != nil {
		h.logger.Error("failed to search stations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search stations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": stations,
		"count":   len(stations),
	})
}

// GetStation handles GET /api/v1/stations/:crs
func (h *StationHandler) GetStation(c *gin.Context) {
	crs := c.Param("crs")
	if crs == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CRS code is required"})
		return
	}

	station, err := h.stationRepo.GetByCRS(c.Request.Context(), crs)
	if err != nil {
		h.logger.Error("failed to get station", zap.Error(err), zap.String("crs", crs))
		c.JSON(http.StatusNotFound, gin.H{"error": "station not found"})
		return
	}

	c.JSON(http.StatusOK, station)
}
