package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	profileProxy *ProfileProxy
	botProxy     *BotProxy
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(profileProxy *ProfileProxy, botProxy *BotProxy) *HealthHandler {
	return &HealthHandler{
		profileProxy: profileProxy,
		botProxy:     botProxy,
	}
}

// Health returns the health status of the API Gateway.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "api-gateway",
		"version": "1.0.0",
	})
}

// Ready checks if the API Gateway is ready to serve requests.
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check backend services
	profileHealthy := h.profileProxy.IsHealthy()
	botHealthy := h.botProxy.IsHealthy()

	if !profileHealthy || !botHealthy {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"service": "api-gateway",
			"details": gin.H{
				"profile_service": profileHealthy,
				"bot_service":     botHealthy,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "api-gateway",
		"details": gin.H{
			"profile_service": profileHealthy,
			"bot_service":     botHealthy,
		},
	})
}
