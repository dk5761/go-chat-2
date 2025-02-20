package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Manager  *Manager
	mu       sync.Mutex
	isClosed bool
}

type Manager struct {
	clients     map[string]*Client // Local cache of active connections
	clientStore *ClientStore       // Redis-based client store
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	logger      *logrus.Logger
	wg          conc.WaitGroup
	serverID    string // Unique identifier for this server instance
}

type MessageType string

const (
	MessageTypeChat   MessageType = "chat"
	MessageTypeTyping MessageType = "typing"
	MessageTypeRead   MessageType = "read"
)

type WebSocketMessage struct {
	Type        MessageType `json:"type"`
	SenderID    string      `json:"sender_id"`
	RecipientID *string     `json:"recipient_id,omitempty"`
	GroupID     *string     `json:"group_id,omitempty"`
	Content     string      `json:"content"`
	Timestamp   time.Time   `json:"timestamp"`
}

func NewManager(logger *logrus.Logger, redisClient *redis.Client) *Manager {
	serverID := uuid.New().String() // Generate unique server ID
	return &Manager{
		clients:     make(map[string]*Client),
		clientStore: NewClientStore(redisClient),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		logger:      logger,
		serverID:    serverID,
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.logger.WithField("server_id", m.serverID).Info("Starting WebSocket manager")

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Shutting down WebSocket manager")
			m.shutdown()
			return

		case client := <-m.register:
			m.mu.Lock()
			m.clients[client.ID] = client
			m.mu.Unlock()

			// Register in Redis
			if err := m.clientStore.AddClient(ctx, client.ID, m.serverID); err != nil {
				m.logger.WithError(err).Error("Failed to register client in Redis")
			}

			m.logger.WithFields(logrus.Fields{
				"client_id":     client.ID,
				"server_id":     m.serverID,
				"total_clients": len(m.clients),
			}).Info("Client connected")

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)

				// Remove from Redis
				if err := m.clientStore.RemoveClient(ctx, client.ID); err != nil {
					m.logger.WithError(err).Error("Failed to remove client from Redis")
				}

				m.logger.WithFields(logrus.Fields{
					"client_id":     client.ID,
					"server_id":     m.serverID,
					"total_clients": len(m.clients),
				}).Info("Client disconnected")
			}
			m.mu.Unlock()

		case message := <-m.broadcast:
			m.mu.RLock()
			for _, client := range m.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					m.mu.RUnlock()
					m.mu.Lock()
					delete(m.clients, client.ID)

					// Remove from Redis
					if err := m.clientStore.RemoveClient(ctx, client.ID); err != nil {
						m.logger.WithError(err).Error("Failed to remove client from Redis")
					}

					m.mu.Unlock()
					m.mu.RLock()
				}
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, client := range m.clients {
		client.mu.Lock()
		if !client.isClosed {
			close(client.Send)
			client.isClosed = true
		}
		client.mu.Unlock()
		client.Conn.Close()
	}

	// Wait for all goroutines to complete
	m.wg.Wait()
}

func (m *Manager) SendToUser(userID string, message []byte) error {
	// First check if user is connected to any server
	connected, err := m.clientStore.IsConnected(context.Background(), userID)
	if err != nil {
		m.logger.WithError(err).Error("Failed to check client connection status")
		return err
	}

	if !connected {
		m.logger.WithField("user_id", userID).Debug("User is offline")
		return nil
	}

	// Check if client is connected to this server
	m.mu.RLock()
	client, exists := m.clients[userID]
	m.mu.RUnlock()

	if !exists {
		m.logger.WithFields(logrus.Fields{
			"user_id":   userID,
			"server_id": m.serverID,
			"connected": connected,
		}).Debug("User is connected to a different server")
		return nil
	}

	select {
	case client.Send <- message:
		return nil
	default:
		return nil // Channel is full or closed
	}
}

func (m *Manager) SendToGroup(groupID string, message []byte, excludeUserID string) {
	m.broadcast <- message
}

func (m *Manager) HandleClient(client *Client) {
	// Register the client first
	m.register <- client

	// Start client handlers in the wait group
	m.wg.Go(func() {
		go client.writePump()
		go client.readPump()
	})
}

func (c *Client) writePump() {
	ticker := time.NewTicker(time.Second * 30)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			c.Manager.logger.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		// Set sender ID from authenticated client
		wsMessage.SenderID = c.ID
		wsMessage.Timestamp = time.Now()

		switch wsMessage.Type {
		case MessageTypeChat:
			// For 1-to-1 chat
			if wsMessage.RecipientID != nil {
				if err := c.Manager.SendToUser(*wsMessage.RecipientID, message); err != nil {
					c.Manager.logger.Errorf("Failed to send message: %v", err)
				}
			}
			// For group chat
			if wsMessage.GroupID != nil {
				c.Manager.SendToGroup(*wsMessage.GroupID, message, c.ID)
			}
		case MessageTypeTyping:
			// Handle typing indicators
			if wsMessage.RecipientID != nil {
				if err := c.Manager.SendToUser(*wsMessage.RecipientID, message); err != nil {
					c.Manager.logger.Errorf("Failed to send typing indicator: %v", err)
				}
			}
		case MessageTypeRead:
			// Handle read receipts
			if wsMessage.RecipientID != nil {
				if err := c.Manager.SendToUser(*wsMessage.RecipientID, message); err != nil {
					c.Manager.logger.Errorf("Failed to send read receipt: %v", err)
				}
			}
		}
	}
}
