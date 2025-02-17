package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/chat-backend/internal/models"
	"github.com/chat-backend/internal/repository"
	"github.com/chat-backend/internal/websocket"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc"
)

type MessageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
	wsManager   *websocket.Manager
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	groupRepo repository.GroupRepository,
	wsManager *websocket.Manager,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		wsManager:   wsManager,
	}
}

type SendMessageInput struct {
	SenderID    uuid.UUID  `json:"sender_id"`
	RecipientID *uuid.UUID `json:"recipient_id,omitempty"`
	GroupID     *uuid.UUID `json:"group_id,omitempty"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	ReplyToID   *uuid.UUID `json:"reply_to_id,omitempty"`
	Attachments []string   `json:"attachments,omitempty"`
}

func (s *MessageService) SendMessage(ctx context.Context, input SendMessageInput) (*models.Message, error) {
	// Validate input
	if input.RecipientID == nil && input.GroupID == nil {
		return nil, errors.New("either recipient_id or group_id must be provided")
	}
	if input.RecipientID != nil && input.GroupID != nil {
		return nil, errors.New("cannot send message to both user and group")
	}

	// Create message
	message := &models.Message{
		ID:          uuid.New(),
		SenderID:    input.SenderID,
		RecipientID: input.RecipientID,
		GroupID:     input.GroupID,
		Content:     input.Content,
		ContentType: input.ContentType,
		Timestamp:   time.Now(),
		ReadBy:      []uuid.UUID{input.SenderID}, // Mark as read by sender
		DeliveredTo: []uuid.UUID{input.SenderID}, // Mark as delivered to sender
		ReplyToID:   input.ReplyToID,
		Attachments: input.Attachments,
	}

	// Use conc.WaitGroup for concurrent operations
	var wg conc.WaitGroup
	var saveErr error
	var deliveryErr error

	// Save message
	wg.Go(func() {
		if err := s.messageRepo.Create(ctx, message); err != nil {
			saveErr = err
		}
	})

	// Handle delivery
	wg.Go(func() {
		if input.RecipientID != nil {
			// Direct message
			if err := s.deliverDirectMessage(ctx, message); err != nil {
				deliveryErr = err
			}
		} else {
			// Group message
			if err := s.deliverGroupMessage(ctx, message); err != nil {
				deliveryErr = err
			}
		}
	})

	wg.Wait()

	if saveErr != nil {
		return nil, saveErr
	}
	if deliveryErr != nil {
		return nil, deliveryErr
	}

	return message, nil
}

func (s *MessageService) deliverDirectMessage(ctx context.Context, message *models.Message) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send to recipient via WebSocket if online
	if err := s.wsManager.SendToUser(*message.RecipientID, messageJSON); err != nil {
		// TODO: Queue for push notification if delivery fails
		return err
	}

	return nil
}

func (s *MessageService) deliverGroupMessage(ctx context.Context, message *models.Message) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send to all group members except sender
	s.wsManager.SendToGroup(*message.GroupID, messageJSON, message.SenderID)
	return nil
}

func (s *MessageService) GetMessage(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	return s.messageRepo.GetByID(ctx, id)
}

func (s *MessageService) GetUserMessages(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Message, error) {
	return s.messageRepo.GetUserMessages(ctx, userID, limit, offset)
}

func (s *MessageService) GetGroupMessages(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]models.Message, error) {
	return s.messageRepo.GetGroupMessages(ctx, groupID, limit, offset)
}

func (s *MessageService) GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit, offset int) ([]models.Message, error) {
	return s.messageRepo.GetConversation(ctx, user1ID, user2ID, limit, offset)
}

func (s *MessageService) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	return s.messageRepo.MarkAsRead(ctx, messageID, userID)
}

func (s *MessageService) MarkAsDelivered(ctx context.Context, messageID, userID uuid.UUID) error {
	return s.messageRepo.MarkAsDelivered(ctx, messageID, userID)
}

func (s *MessageService) UpdateMessage(ctx context.Context, message *models.Message) error {
	return s.messageRepo.Update(ctx, message)
}

func (s *MessageService) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return s.messageRepo.Delete(ctx, id)
}
