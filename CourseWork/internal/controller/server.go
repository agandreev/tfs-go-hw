package controller

import (
	"context"
	"github.com/agandreev/tfs-go-hw/CourseWork/internal/handlers"
	"net/http"
	"time"
)

// Server represents http server structure with handler's implementations.
type Server struct {
	httpServer *http.Server
	handler    handlers.Handler
}

// NewServer creates Server pointer.
func NewServer(handler handlers.Handler) *Server {
	return &Server{handler: handler}
}

// Run runs http server on chosen port with handlers from Server and sets timeouts.
func (s *Server) Run(port string) error {
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      s.handler.InitRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops all handler's goroutines and http server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.handler.Trader.ShutDown()
	return s.httpServer.Shutdown(ctx)
}
