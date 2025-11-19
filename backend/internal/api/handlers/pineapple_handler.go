package handlers

import (
	"net/http"
	"strconv"

	"github.com/dcvdiego/trainpain-backend/internal/domain"
	"github.com/dcvdiego/trainpain-backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PineappleHandler struct {
	service *service.PineappleService
	logger  *zap.Logger
}

func NewPineappleHandler(service *service.PineappleService, logger *zap.Logger) *PineappleHandler {
	return &PineappleHandler{
		service: service,
		logger:  logger,
	}
}

// GetAllClasses returns all classes with their schedules
func (h *PineappleHandler) GetAllClasses(c *gin.Context) {
	classes, err := h.service.GetAllClasses(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get classes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"classes": classes})
}

// GetClassesByDay returns classes for a specific day of the week
func (h *PineappleHandler) GetClassesByDay(c *gin.Context) {
	dayStr := c.Param("day")
	day, err := strconv.Atoi(dayStr)
	if err != nil || day < 0 || day > 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid day of week (must be 0-6)"})
		return
	}

	classes, err := h.service.GetClassesByDay(c.Request.Context(), day)
	if err != nil {
		h.logger.Error("failed to get classes by day", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"classes": classes})
}

// Subscribe creates a new subscription
func (h *PineappleHandler) Subscribe(c *gin.Context) {
	var req struct {
		ScheduleID            int    `json:"schedule_id" binding:"required"`
		PushEndpoint          string `json:"push_endpoint" binding:"required"`
		PushP256dh            string `json:"push_p256dh" binding:"required"`
		PushAuth              string `json:"push_auth" binding:"required"`
		NotificationThreshold int    `json:"notification_threshold"`
		Notify2hBefore        bool   `json:"notify_2h_before"`
		Notify1hBefore        bool   `json:"notify_1h_before"`
		Notify30mBefore       bool   `json:"notify_30m_before"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default threshold if not provided
	if req.NotificationThreshold == 0 {
		req.NotificationThreshold = 5
	}

	sub := &domain.PineappleSubscription{
		ScheduleID:            req.ScheduleID,
		PushEndpoint:          req.PushEndpoint,
		PushP256dh:            req.PushP256dh,
		PushAuth:              req.PushAuth,
		NotificationThreshold: req.NotificationThreshold,
		Notify2hBefore:        req.Notify2hBefore,
		Notify1hBefore:        req.Notify1hBefore,
		Notify30mBefore:       req.Notify30mBefore,
		IsActive:              true,
	}

	err := h.service.Subscribe(c.Request.Context(), sub)
	if err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subscription": sub})
}

// Unsubscribe removes a subscription
func (h *PineappleHandler) Unsubscribe(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	err = h.service.Unsubscribe(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to unsubscribe", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed successfully"})
}

// GetUserSubscriptions returns all subscriptions for a user
func (h *PineappleHandler) GetUserSubscriptions(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint parameter required"})
		return
	}

	subscriptions, err := h.service.GetUserSubscriptions(c.Request.Context(), endpoint)
	if err != nil {
		h.logger.Error("failed to get subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

// ScrapeClasses triggers a manual scrape of classes
func (h *PineappleHandler) ScrapeClasses(c *gin.Context) {
	err := h.service.ScrapeAllClasses(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to scrape classes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scrape classes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "scraping completed successfully"})
}
