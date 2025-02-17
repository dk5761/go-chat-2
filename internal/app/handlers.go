package app

import (
	"github.com/chat-backend/internal/api"
	"github.com/chat-backend/internal/api/middleware"
)

type handlers struct {
	userHandler    *api.UserHandler
	groupHandler   *api.GroupHandler
	messageHandler *api.MessageHandler
	wsHandler      *api.WebSocketHandler
	authMiddleware *middleware.AuthMiddleware
	healthHandler  *api.HealthHandler
}

func initHandlers(services *services) *handlers {
	return &handlers{
		userHandler:    api.NewUserHandler(services.userService),
		groupHandler:   api.NewGroupHandler(services.groupService),
		messageHandler: api.NewMessageHandler(services.messageService),
		wsHandler:      api.NewWebSocketHandler(services.wsManager, services.userService, services.messageService),
		authMiddleware: middleware.NewAuthMiddleware(services.userService),
		healthHandler:  api.NewHealthHandler(),
	}
}
