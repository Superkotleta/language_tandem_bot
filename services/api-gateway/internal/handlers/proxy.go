package handlers

import (
	"context"
	"net/http"
	"time"

	"api-gateway/internal/proxy"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProfileProxy handles requests to the Profile service.
type ProfileProxy struct {
	proxy  *proxy.Proxy
	logger *zap.Logger
}

// NewProfileProxy creates a new profile proxy.
func NewProfileProxy(proxy *proxy.Proxy, logger *zap.Logger) *ProfileProxy {
	return &ProfileProxy{
		proxy:  proxy,
		logger: logger,
	}
}

// Forward forwards requests to the Profile service.
func (p *ProfileProxy) Forward(c *gin.Context) {
	// Keep the full path for profile service (including /api/v1)
	path := c.Request.URL.Path

	resp, err := p.proxy.ForwardRequest(c.Request.Context(), c.Request, path)
	if err != nil {
		p.logger.Error("Failed to forward request to profile service",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Service unavailable",
			"message": "Profile service is not available",
		})
		return
	}
	defer resp.Body.Close()

	// Forward response
	if err := p.proxy.ForwardResponse(c.Writer, resp); err != nil {
		p.logger.Error("Failed to forward response from profile service",
			zap.String("path", path),
			zap.Error(err),
		)
	}
}

// IsHealthy checks if the Profile service is healthy.
func (p *ProfileProxy) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.proxy.HealthCheck(ctx)
	return err == nil
}

// BotProxy handles requests to the Bot service.
type BotProxy struct {
	proxy  *proxy.Proxy
	logger *zap.Logger
}

// NewBotProxy creates a new bot proxy.
func NewBotProxy(proxy *proxy.Proxy, logger *zap.Logger) *BotProxy {
	return &BotProxy{
		proxy:  proxy,
		logger: logger,
	}
}

// Forward forwards requests to the Bot service.
func (p *BotProxy) Forward(c *gin.Context) {
	// Keep the full path for bot service
	path := c.Request.URL.Path

	resp, err := p.proxy.ForwardRequest(c.Request.Context(), c.Request, path)
	if err != nil {
		p.logger.Error("Failed to forward request to bot service",
			zap.String("path", path),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Service unavailable",
			"message": "Bot service is not available",
		})
		return
	}
	defer resp.Body.Close()

	// Forward response
	if err := p.proxy.ForwardResponse(c.Writer, resp); err != nil {
		p.logger.Error("Failed to forward response from bot service",
			zap.String("path", path),
			zap.Error(err),
		)
	}
}

// IsHealthy checks if the Bot service is healthy.
func (p *BotProxy) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.proxy.HealthCheck(ctx)
	return err == nil
}
