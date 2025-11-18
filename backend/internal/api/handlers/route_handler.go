package handlers

import (
	"net/http"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/dcvdiego/trainpain-backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RouteHandler struct {
	routeService *service.RouteService
	logger       *zap.Logger
}

func NewRouteHandler(routeService *service.RouteService, logger *zap.Logger) *RouteHandler {
	return &RouteHandler{
		routeService: routeService,
		logger:       logger,
	}
}

// GetReliability handles POST /api/v1/routes/reliability
func (h *RouteHandler) GetReliability(c *gin.Context) {
	var req domain.RouteReliabilityQuery

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("reliability query received",
		zap.String("origin", req.OriginCRS),
		zap.String("destination", req.DestinationCRS),
		zap.Int("analysis_days", req.AnalysisDays),
	)

	response, err := h.routeService.GetReliability(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to get reliability", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute reliability metrics"})
		return
	}

	c.JSON(http.StatusOK, response)
}
