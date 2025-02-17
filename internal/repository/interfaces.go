package repository

import (
	"context"

	"github.com/chat-backend/internal/models"
	"github.com/google/uuid"
)

// UserRepository handles all user-related database operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
}

// GroupRepository handles all group-related database operations
type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, groupID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	GetMembers(ctx context.Context, groupID uuid.UUID) ([]models.GroupMember, error)
	GetUserGroups(ctx context.Context, userID uuid.UUID) ([]models.Group, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role string) error
}

// MessageRepository handles all message-related database operations
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetUserMessages(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Message, error)
	GetGroupMessages(ctx context.Context, groupID uuid.UUID, limit int, offset int) ([]models.Message, error)
	GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit int, offset int) ([]models.Message, error)
	MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error
	MarkAsDelivered(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// StatusRepository handles all user status related operations
type StatusRepository interface {
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
	GetStatus(ctx context.Context, userID uuid.UUID) (string, error)
	GetMultiStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
}
