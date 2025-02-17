package api

import (
	"errors"
	"net/http"

	"github.com/chat-backend/internal/service"
	wsmanager "github.com/chat-backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	wsManager      *wsmanager.Manager
	userService    *service.UserService
	messageService *service.MessageService
}

func NewWebSocketHandler(
	wsManager *wsmanager.Manager,
	userService *service.UserService,
	messageService *service.MessageService,
) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager:      wsManager,
		userService:    userService,
		messageService: messageService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin check
		return true
	},
}

func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// Get user ID from auth token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not upgrade connection"})
		return
	}

	// Create new client
	client := &wsmanager.Client{
		ID:      userID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: h.wsManager,
	}

	// Register client
	h.wsManager.HandleClient(client)

	// Update user status to online
	if err := h.userService.UpdateStatus(c.Request.Context(), userID, "online"); err != nil {
		// Log error but don't fail the connection
		// TODO: Add proper logging
	}
}

func (h *WebSocketHandler) getUserIDFromToken(c *gin.Context) (uuid.UUID, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
		if token == "" {
			return uuid.Nil, errors.New("no token provided")
		}
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	return h.userService.ValidateToken(token)
}

// RegisterRoutes registers the WebSocket routes
func (h *WebSocketHandler) RegisterRoutes(router *gin.RouterGroup) {
	ws := router.Group("/ws")
	{
		ws.GET("", h.HandleConnection)
	}
}
