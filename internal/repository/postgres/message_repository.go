package postgres

import (
	"context"
	"time"

	"github.com/chat-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *messageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetUserMessages(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where("sender_id = ? OR recipient_id = ?", userID, userID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetGroupMessages(ctx context.Context, groupID uuid.UUID, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetMessagesBetween(ctx context.Context, userID1, userID2 uuid.UUID, limit int64, before time.Time) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.WithContext(ctx).
		Where(
			r.db.Where("sender_id = ? AND recipient_id = ?", userID1, userID2).
				Or("sender_id = ? AND recipient_id = ?", userID2, userID1),
		).
		Where("timestamp < ?", before).
		Order("timestamp DESC").
		Limit(int(limit)).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("read_by", gorm.Expr("array_append(read_by, ?)", userID.String())).
		Error
}

func (r *messageRepository) MarkAsDelivered(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", messageID).
		Update("delivered_to", gorm.Expr("array_append(delivered_to, ?)", userID.String())).
		Error
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, id).Error
}
