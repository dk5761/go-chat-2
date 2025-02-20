package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chat-backend/internal/api"
	"github.com/chat-backend/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	logger     *logrus.Logger
}

func NewServer(
	logger *logrus.Logger,
	userHandler *api.UserHandler,
	groupHandler *api.GroupHandler,
	messageHandler *api.MessageHandler,
	wsHandler *api.WebSocketHandler,
	authMiddleware *middleware.AuthMiddleware,
	healthHandler *api.HealthHandler,
) *Server {
	router := gin.Default()

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		userHandler.RegisterRoutes(v1)
		healthHandler.RegisterRoutes(v1)

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			groupHandler.RegisterRoutes(protected)
			messageHandler.RegisterRoutes(protected)
			wsHandler.RegisterRoutes(protected)
		}
	}

	return &Server{
		router: router,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%s", viper.GetString("server.port")),
			Handler: router,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
