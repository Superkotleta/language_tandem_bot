package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	telegramHealthy bool
	discordHealthy  bool
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		telegramHealthy: true,  // TODO: implement actual health checks
		discordHealthy:  false, // Discord not implemented yet
	}
}

// Health returns the health status of the Bot service.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "bot",
		"version": "1.0.0",
	})
}

// Ready checks if the Bot service is ready to serve requests.
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check bot services
	if !h.telegramHealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"service": "bot",
			"details": gin.H{
				"telegram_bot": h.telegramHealthy,
				"discord_bot":  h.discordHealthy,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "bot",
		"details": gin.H{
			"telegram_bot": h.telegramHealthy,
			"discord_bot":  h.discordHealthy,
		},
	})
}
