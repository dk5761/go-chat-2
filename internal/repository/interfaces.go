package repository

import (
	"context"
	"time"

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
	GetByID(ctx context.Context, id string) (*models.Message, error)
	GetUserMessages(ctx context.Context, userID string, limit int, offset int) ([]models.Message, error)
	GetGroupMessages(ctx context.Context, groupID string, limit int, offset int) ([]models.Message, error)
	GetMessagesBetween(ctx context.Context, userID1, userID2 string, limit int64, before time.Time) ([]*models.Message, error)
	MarkAsRead(ctx context.Context, messageID string, userID string) error
	MarkAsDelivered(ctx context.Context, messageID string, userID string) error
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id string) error
}

// StatusRepository handles all user status related operations
type StatusRepository interface {
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
	GetStatus(ctx context.Context, userID uuid.UUID) (string, error)
	GetMultiStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
}
