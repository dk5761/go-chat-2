package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	wsClientKeyPrefix = "ws:client:"
	wsClientTTL       = 24 * time.Hour
)

type ClientStore struct {
	redis *redis.Client
}

type ClientInfo struct {
	UserID    string    `json:"user_id"`
	ServerID  string    `json:"server_id"` // For identifying which server instance the client is connected to
	LastSeen  time.Time `json:"last_seen"`
	Connected bool      `json:"connected"`
}

func NewClientStore(redisClient *redis.Client) *ClientStore {
	return &ClientStore{
		redis: redisClient,
	}
}

func (s *ClientStore) AddClient(ctx context.Context, userID, serverID string) error {
	key := fmt.Sprintf("%s%s", wsClientKeyPrefix, userID)
	info := ClientInfo{
		UserID:    userID,
		ServerID:  serverID,
		LastSeen:  time.Now(),
		Connected: true,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal client info: %w", err)
	}

	return s.redis.Set(ctx, key, data, wsClientTTL).Err()
}

func (s *ClientStore) RemoveClient(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", wsClientKeyPrefix, userID)
	return s.redis.Del(ctx, key).Err()
}

func (s *ClientStore) GetClient(ctx context.Context, userID string) (*ClientInfo, error) {
	key := fmt.Sprintf("%s%s", wsClientKeyPrefix, userID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client info: %w", err)
	}

	var info ClientInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client info: %w", err)
	}

	return &info, nil
}

func (s *ClientStore) UpdateLastSeen(ctx context.Context, userID string) error {
	info, err := s.GetClient(ctx, userID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("client not found")
	}

	info.LastSeen = time.Now()
	return s.AddClient(ctx, userID, info.ServerID)
}

func (s *ClientStore) IsConnected(ctx context.Context, userID string) (bool, error) {
	info, err := s.GetClient(ctx, userID)
	if err != nil {
		return false, err
	}
	if info == nil {
		return false, nil
	}
	return info.Connected, nil
}

func (s *ClientStore) GetAllClients(ctx context.Context) ([]ClientInfo, error) {
	pattern := fmt.Sprintf("%s*", wsClientKeyPrefix)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get client keys: %w", err)
	}

	var clients []ClientInfo
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var info ClientInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		clients = append(clients, info)
	}

	return clients, nil
}
