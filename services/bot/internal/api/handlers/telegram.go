package handlers

import (
	"net/http"

	"language-exchange-bot/internal/adapters/telegram"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TelegramHandler handles Telegram webhook requests.
type TelegramHandler struct {
	bot    *telegram.TelegramBot
	logger *zap.Logger
}

// NewTelegramHandler creates a new Telegram handler.
func NewTelegramHandler(bot *telegram.TelegramBot, logger *zap.Logger) *TelegramHandler {
	return &TelegramHandler{
		bot:    bot,
		logger: logger,
	}
}

// Webhook handles incoming Telegram webhook updates.
func (h *TelegramHandler) Webhook(c *gin.Context) {
	var update telegram.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		h.logger.Error("Failed to bind webhook update",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}

	// Process the update
	if err := h.bot.ProcessUpdate(&update); err != nil {
		h.logger.Error("Failed to process update",
			zap.Error(err),
			zap.Int("update_id", update.UpdateID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process update",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// SetWebhook sets the Telegram webhook URL.
func (h *TelegramHandler) SetWebhook(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.bot.SetWebhook(req.URL); err != nil {
		h.logger.Error("Failed to set webhook",
			zap.Error(err),
			zap.String("url", req.URL),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to set webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "webhook set successfully",
		"url":    req.URL,
	})
}

// GetWebhookInfo gets the current webhook information.
func (h *TelegramHandler) GetWebhookInfo(c *gin.Context) {
	info, err := h.bot.GetWebhookInfo()
	if err != nil {
		h.logger.Error("Failed to get webhook info",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get webhook info",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// SendMessage sends a message to a user.
func (h *TelegramHandler) SendMessage(c *gin.Context) {
	var req struct {
		ChatID    int64  `json:"chat_id" binding:"required"`
		Text      string `json:"text" binding:"required"`
		ParseMode string `json:"parse_mode,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	message := telegram.SendMessageRequest{
		ChatID:    req.ChatID,
		Text:      req.Text,
		ParseMode: req.ParseMode,
	}

	if err := h.bot.SendMessage(&message); err != nil {
		h.logger.Error("Failed to send message",
			zap.Error(err),
			zap.Int64("chat_id", req.ChatID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "message sent successfully",
	})
}
