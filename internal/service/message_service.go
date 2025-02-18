package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/conc"

	"github.com/chat-backend/internal/models"
	"github.com/chat-backend/internal/repository"
	"github.com/chat-backend/internal/websocket"
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
	SenderID    string   `json:"sender_id"`
	RecipientID *string  `json:"recipient_id,omitempty"`
	GroupID     *string  `json:"group_id,omitempty"`
	Content     string   `json:"content"`
	ContentType string   `json:"content_type"`
	ReplyToID   *string  `json:"reply_to_id,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

func (s *MessageService) SendMessage(ctx context.Context, input SendMessageInput) (*models.Message, error) {
	// Validate input
	if input.RecipientID == nil && input.GroupID == nil {
		return nil, errors.New("either recipient_id or group_id must be provided")
	}
	if input.RecipientID != nil && input.GroupID != nil {
		return nil, errors.New("cannot send message to both user and group")
	}

	// Convert string IDs to UUIDs
	senderUUID, err := uuid.Parse(input.SenderID)
	if err != nil {
		return nil, errors.New("invalid sender ID")
	}

	var recipientUUID *uuid.UUID
	if input.RecipientID != nil {
		parsed, err := uuid.Parse(*input.RecipientID)
		if err != nil {
			return nil, errors.New("invalid recipient ID")
		}
		recipientUUID = &parsed
	}

	var groupUUID *uuid.UUID
	if input.GroupID != nil {
		parsed, err := uuid.Parse(*input.GroupID)
		if err != nil {
			return nil, errors.New("invalid group ID")
		}
		groupUUID = &parsed
	}

	var replyToUUID *uuid.UUID
	if input.ReplyToID != nil {
		parsed, err := uuid.Parse(*input.ReplyToID)
		if err != nil {
			return nil, errors.New("invalid reply-to ID")
		}
		replyToUUID = &parsed
	}

	// Create message
	message := &models.Message{
		ID:          uuid.New(),
		SenderID:    senderUUID,
		RecipientID: recipientUUID,
		GroupID:     groupUUID,
		Content:     input.Content,
		ContentType: input.ContentType,
		Timestamp:   time.Now(),
		ReadBy:      []string{input.SenderID},
		DeliveredTo: []string{input.SenderID},
		ReplyToID:   replyToUUID,
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
	if err := s.wsManager.SendToUser(message.RecipientID.String(), messageJSON); err != nil {
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
	s.wsManager.SendToGroup(message.GroupID.String(), messageJSON, message.SenderID.String())
	return nil
}

func (s *MessageService) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	messageID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid message ID")
	}
	return s.messageRepo.GetByID(ctx, messageID)
}

func (s *MessageService) GetUserMessages(ctx context.Context, userID string, limit, offset int) ([]models.Message, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	return s.messageRepo.GetUserMessages(ctx, userUUID, limit, offset)
}

func (s *MessageService) GetGroupMessages(ctx context.Context, groupID string, limit, offset int) ([]models.Message, error) {
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}
	return s.messageRepo.GetGroupMessages(ctx, groupUUID, limit, offset)
}

func (s *MessageService) GetConversation(ctx context.Context, user1ID, user2ID string, limit, offset int) ([]models.Message, error) {
	user1UUID, err := uuid.Parse(user1ID)
	if err != nil {
		return nil, errors.New("invalid user1 ID")
	}
	user2UUID, err := uuid.Parse(user2ID)
	if err != nil {
		return nil, errors.New("invalid user2 ID")
	}

	messages, err := s.messageRepo.GetMessagesBetween(ctx, user1UUID, user2UUID, int64(limit), time.Now())
	if err != nil {
		return nil, err
	}
	result := make([]models.Message, len(messages))
	for i, msg := range messages {
		result[i] = *msg
	}
	return result, nil
}

func (s *MessageService) MarkAsRead(ctx context.Context, messageID string, userID string) error {
	msgUUID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	return s.messageRepo.MarkAsRead(ctx, msgUUID, userUUID)
}

func (s *MessageService) MarkAsDelivered(ctx context.Context, messageID string, userID string) error {
	msgUUID, err := uuid.Parse(messageID)
	if err != nil {
		return errors.New("invalid message ID")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	return s.messageRepo.MarkAsDelivered(ctx, msgUUID, userUUID)
}

func (s *MessageService) UpdateMessage(ctx context.Context, message *models.Message) error {
	return s.messageRepo.Update(ctx, message)
}

func (s *MessageService) DeleteMessage(ctx context.Context, id string) error {
	messageID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid message ID")
	}
	return s.messageRepo.Delete(ctx, messageID)
}
