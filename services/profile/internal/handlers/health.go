package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db interface {
		Ping(ctx context.Context) error
	}
}

func NewHealthHandler(db interface {
	Ping(ctx context.Context) error
}) *HealthHandler {
	return &HealthHandler{db: db}
}

// @Router /healthz [get].
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "profile",
	})
}

// @Router /readyz [get].
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check database connection
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "Service not ready",
			Message: "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "profile",
	})
}
