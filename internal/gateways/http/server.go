package http

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/usecase"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host       string
	port       uint16
	router     *gin.Engine
	httpServer *http.Server
	wsHandler  *WebSocketHandler
}

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	wsHandler := NewWebSocketHandler(useCases)
	setupRouter(r, useCases, wsHandler)

	s := &Server{
		router:    r,
		host:      "localhost",
		port:      8080,
		wsHandler: wsHandler,
	}

	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %s\n", err)
			quit <- syscall.SIGTERM
		}
	}()
	select {
	case <-quit:
		log.Println("Shutdown signal received")
	case <-ctx.Done():
		log.Println("Context canceled")
	}

	return s.Shutdown(context.Background())
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Println("Closing WebSocket connections...")
	if err := s.wsHandler.Shutdown(); err != nil {
		log.Printf("WebSocket shutdown error: %v", err)
	}

	log.Println("Shutting down HTTP server...")
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
