// Package server provides HTTP server implementation for administrative API
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"language-exchange-bot/internal/adapters/telegram"
	"language-exchange-bot/internal/core"
	"language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/models"
	docs "language-exchange-bot/internal/server/docs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// AdminServer provides REST API for administrative operations and webhook handling
type AdminServer struct {
	port        string
	botService  *core.BotService
	handler     *telegram.TelegramHandler
	webhookMode bool
	server      *http.Server
}

// New creates a new admin HTTP server
func New(port string, botService *core.BotService, handler *telegram.TelegramHandler) *AdminServer {
	return NewWithWebhook(port, botService, handler, false)
}

// NewWithWebhook creates a new admin HTTP server with webhook support
func NewWithWebhook(port string, botService *core.BotService, handler *telegram.TelegramHandler, webhookMode bool) *AdminServer {
	r := mux.NewRouter()

	s := &AdminServer{
		port:        port,
		botService:  botService,
		handler:     handler,
		webhookMode: webhookMode,
	}

	// Initialize swagger docs
	docs.SwaggerInfo.Host = "localhost:" + port
	docs.SwaggerInfo.BasePath = "/api/v1"

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(s.corsMiddleware)
	api.Use(s.authMiddleware)

	// Health check
	r.HandleFunc("/healthz", s.handleHealth).Methods("GET")
	r.HandleFunc("/readyz", s.handleReady).Methods("GET")

	// Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Static files and navigation page
	r.HandleFunc("/", s.handleNavigation).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Telegram webhook endpoint (only if webhook mode is enabled)
	if webhookMode && handler != nil {
		r.HandleFunc("/webhook/telegram/{token}", s.handleTelegramWebhook).Methods("POST")
	}

	// API endpoints
	api.HandleFunc("/stats", s.handleGetStats).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", s.handleGetUser).Methods("GET")
	api.HandleFunc("/users", s.handleGetUsers).Methods("GET").Queries("limit", "{limit:[0-9]+}", "offset", "{offset:[0-9]+}")
	api.HandleFunc("/feedback/unprocessed", s.handleGetUnprocessedFeedback).Methods("GET")
	api.HandleFunc("/feedback/{id:[0-9]+}/process", s.handleProcessFeedback).Methods("POST")
	api.HandleFunc("/rate-limits/stats", s.handleGetRateLimitStats).Methods("GET")
	api.HandleFunc("/cache/stats", s.handleGetCacheStats).Methods("GET")
	api.HandleFunc("/webhook/status", s.handleGetWebhookStatus).Methods("GET")
	api.HandleFunc("/webhook/setup", s.handleSetupWebhook).Methods("POST")
	api.HandleFunc("/webhook/remove", s.handleRemoveWebhook).Methods("POST")

	s.server = &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return s
}

// Start starts the admin HTTP server
func (s *AdminServer) Start() error {
	fmt.Printf("Admin API server starting on port %s\n", s.port)
	return s.server.ListenAndServe()
}

// Stop stops the admin HTTP server
func (s *AdminServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// corsMiddleware adds CORS headers
func (s *AdminServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware provides basic authentication for admin endpoints
func (s *AdminServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement proper authentication
		// For now, just check for a simple header
		if auth := r.Header.Get("X-Admin-Key"); auth != "admin-secret-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth provides health check endpoint
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (s *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleReady provides readiness check endpoint
// @Summary Readiness check
// @Description Check if the service is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /readyz [get]
func (s *AdminServer) handleReady(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetStats returns general statistics
// @Summary Get general statistics
// @Description Retrieve general bot statistics
// @Tags statistics
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/stats [get]
func (s *AdminServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"version":      "3.0.0",
		"service":      "language-exchange-bot",
		"uptime":       "simulated uptime data", // TODO: Add real uptime tracking
		"active_users": 0,                       // TODO: Add real user count
		"total_users":  0,                       // TODO: Add real user count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleGetUser returns user information
// @Summary Get user by ID
// @Description Retrieve user information by Telegram ID
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Telegram User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (s *AdminServer) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := s.botService.GetCachedUser(userID)
	if err != nil {
		if customErr, ok := err.(*errors.CustomError); ok && customErr.Type == errors.ErrorTypeDatabase {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleGetUsers returns list of users with pagination
// @Summary Get users list
// @Description Retrieve paginated list of users
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Number of users to return" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (s *AdminServer) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement pagination
	// For now, return empty list
	response := map[string]interface{}{
		"users":  []models.User{},
		"total":  0,
		"limit":  50,
		"offset": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetUnprocessedFeedback returns unprocessed feedback
// @Summary Get unprocessed feedback
// @Description Retrieve all unprocessed user feedback
// @Tags feedback
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/feedback/unprocessed [get]
func (s *AdminServer) handleGetUnprocessedFeedback(w http.ResponseWriter, r *http.Request) {
	feedback, err := s.botService.GetAllUnprocessedFeedback()
	if err != nil {
		http.Error(w, "Failed to get feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedback)
}

// handleProcessFeedback marks feedback as processed
// @Summary Process feedback
// @Description Mark feedback as processed with optional response
// @Tags feedback
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Feedback ID"
// @Param request body map[string]string true "Processing request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/feedback/{id}/process [post]
func (s *AdminServer) handleProcessFeedback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feedbackIDStr := vars["id"]

	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		http.Error(w, "Invalid feedback ID", http.StatusBadRequest)
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	adminResponse, exists := req["admin_response"]
	if !exists {
		adminResponse = ""
	}

	err = s.botService.MarkFeedbackProcessed(feedbackID, adminResponse)
	if err != nil {
		if customErr, ok := err.(*errors.CustomError); ok && customErr.Type == errors.ErrorTypeDatabase {
			http.Error(w, "Feedback not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to process feedback", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": "processed",
		"id":     feedbackIDStr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetRateLimitStats returns rate limiting statistics
// @Summary Get rate limit statistics
// @Description Retrieve rate limiting statistics
// @Tags monitoring
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rate-limits/stats [get]
func (s *AdminServer) handleGetRateLimitStats(w http.ResponseWriter, r *http.Request) {
	if s.handler == nil {
		http.Error(w, "Handler not available", http.StatusServiceUnavailable)
		return
	}

	stats := s.handler.GetRateLimiterStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleGetCacheStats returns cache statistics
// @Summary Get cache statistics
// @Description Retrieve cache performance statistics
// @Tags monitoring
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cache/stats [get]
func (s *AdminServer) handleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := s.botService.GetCacheStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleNavigation serves the main navigation page
// @Summary Get navigation page
// @Description Returns the main navigation dashboard
// @Tags navigation
// @Produce html
// @Success 200 {string} string "HTML page"
// @Router / [get]
func (s *AdminServer) handleNavigation(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

// handleTelegramWebhook handles Telegram webhook requests
// @Summary Handle Telegram webhook
// @Description Receives and processes webhook updates from Telegram
// @Tags webhook
// @Accept json
// @Param token path string true "Bot token"
// @Param update body object true "Telegram update JSON"
// @Success 200 {string} string "OK"
// @Router /webhook/telegram/{token} [post]
func (s *AdminServer) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	// Basic token validation (should be improved with proper secret)
	// In production, this should validate against a stored secret
	if token == "" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse Telegram update
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Failed to parse update", http.StatusBadRequest)
		return
	}

	// Process update asynchronously to avoid blocking webhook response
	go func(upd tgbotapi.Update) {
		if s.handler != nil {
			if err := s.handler.HandleUpdate(upd); err != nil {
				log.Printf("Error handling webhook update: %v", err)
			}
		}
	}(update)

	// Respond immediately to acknowledge receipt
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleGetWebhookStatus returns webhook configuration status
// @Summary Get webhook status
// @Description Returns current webhook configuration and status
// @Tags webhook
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhook/status [get]
func (s *AdminServer) handleGetWebhookStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"webhook_mode_enabled": s.webhookMode,
		"handler_configured":   s.handler != nil,
		"service_available":    s.handler != nil && s.handler.GetService() != nil,
		"bot_api_available":    s.handler != nil && s.handler.GetBotAPI() != nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleSetupWebhook sets up webhook for the bot
// @Summary Setup webhook
// @Description Configures webhook for Telegram bot
// @Tags webhook
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body map[string]string true "Webhook setup parameters"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhook/setup [post]
func (s *AdminServer) handleSetupWebhook(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	webhookURL, exists := req["webhook_url"]
	if !exists || webhookURL == "" {
		http.Error(w, "webhook_url is required", http.StatusBadRequest)
		return
	}

	// This would need access to the bot instance to setup webhook
	// For now, return success with note
	result := map[string]interface{}{
		"status":      "webhook_setup_requested",
		"webhook_url": webhookURL,
		"note":        "Webhook setup requires bot instance access. Use environment variables for initial setup.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleRemoveWebhook removes webhook configuration
// @Summary Remove webhook
// @Description Removes webhook configuration from Telegram bot
// @Tags webhook
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhook/remove [post]
func (s *AdminServer) handleRemoveWebhook(w http.ResponseWriter, r *http.Request) {
	// This would need access to the bot instance to remove webhook
	result := map[string]interface{}{
		"status": "webhook_removal_requested",
		"note":   "Webhook removal requires bot instance access. Use environment variables to switch modes.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
