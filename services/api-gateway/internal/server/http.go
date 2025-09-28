package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server.
type Server struct {
	port   string
	router *gin.Engine
	srv    *http.Server
}

// New creates a new HTTP server.
func New(port string, router *gin.Engine) *Server {
	return &Server{
		port:   port,
		router: router,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.srv = &http.Server{
		Addr:              ":" + s.port,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}
