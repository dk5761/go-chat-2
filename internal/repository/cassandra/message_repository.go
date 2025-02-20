package cassandra

import (
	"context"
	"time"

	"github.com/chat-backend/internal/models"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type messageRepository struct {
	session *gocql.Session
}

func NewMessageRepository(session *gocql.Session) *messageRepository {
	return &messageRepository{session: session}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.session.Query(`
		INSERT INTO messages (
			id, sender_id, recipient_id, group_id, content, content_type,
			timestamp, read_by, delivered_to, reply_to_id, attachments,
			is_edited, edit_timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		message.ID,
		message.SenderID,
		message.RecipientID,
		message.GroupID,
		message.Content,
		message.ContentType,
		message.Timestamp,
		message.ReadBy,
		message.DeliveredTo,
		message.ReplyToID,
		message.Attachments,
		message.IsEdited,
		message.EditTimestamp,
	).WithContext(ctx).Exec()
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.session.Query(`
		SELECT id, sender_id, recipient_id, group_id, content, content_type,
		       timestamp, read_by, delivered_to, reply_to_id, attachments,
		       is_edited, edit_timestamp
		FROM messages
		WHERE id = ?
		LIMIT 1`,
		id,
	).WithContext(ctx).Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.GroupID,
		&message.Content,
		&message.ContentType,
		&message.Timestamp,
		&message.ReadBy,
		&message.DeliveredTo,
		&message.ReplyToID,
		&message.Attachments,
		&message.IsEdited,
		&message.EditTimestamp,
	)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetUserMessages(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	iter := r.session.Query(`
		SELECT id, sender_id, recipient_id, group_id, content, content_type,
		       timestamp, read_by, delivered_to, reply_to_id, attachments,
		       is_edited, edit_timestamp
		FROM messages
		WHERE recipient_id = ?
		ORDER BY timestamp DESC
		LIMIT ?`,
		userID, limit,
	).WithContext(ctx).PageSize(limit).PageState([]byte{}).Iter()

	var message models.Message
	for iter.Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.GroupID,
		&message.Content,
		&message.ContentType,
		&message.Timestamp,
		&message.ReadBy,
		&message.DeliveredTo,
		&message.ReplyToID,
		&message.Attachments,
		&message.IsEdited,
		&message.EditTimestamp,
	) {
		messages = append(messages, message)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) GetGroupMessages(ctx context.Context, groupID uuid.UUID, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	iter := r.session.Query(`
		SELECT id, sender_id, recipient_id, group_id, content, content_type,
		       timestamp, read_by, delivered_to, reply_to_id, attachments,
		       is_edited, edit_timestamp
		FROM messages
		WHERE group_id = ?
		ORDER BY timestamp DESC
		LIMIT ?`,
		groupID, limit,
	).WithContext(ctx).PageSize(limit).PageState([]byte{}).Iter()

	var message models.Message
	for iter.Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.GroupID,
		&message.Content,
		&message.ContentType,
		&message.Timestamp,
		&message.ReadBy,
		&message.DeliveredTo,
		&message.ReplyToID,
		&message.Attachments,
		&message.IsEdited,
		&message.EditTimestamp,
	) {
		messages = append(messages, message)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) GetConversation(ctx context.Context, user1ID, user2ID uuid.UUID, limit int, offset int) ([]models.Message, error) {
	var messages []models.Message
	iter := r.session.Query(`
		SELECT id, sender_id, recipient_id, group_id, content, content_type,
		       timestamp, read_by, delivered_to, reply_to_id, attachments,
		       is_edited, edit_timestamp
		FROM messages
		WHERE (sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)
		ORDER BY timestamp DESC
		LIMIT ?`,
		user1ID, user2ID, user2ID, user1ID, limit,
	).WithContext(ctx).PageSize(limit).PageState([]byte{}).Iter()

	var message models.Message
	for iter.Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.GroupID,
		&message.Content,
		&message.ContentType,
		&message.Timestamp,
		&message.ReadBy,
		&message.DeliveredTo,
		&message.ReplyToID,
		&message.Attachments,
		&message.IsEdited,
		&message.EditTimestamp,
	) {
		messages = append(messages, message)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) MarkAsRead(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	return r.session.Query(`
		UPDATE messages
		SET read_by = read_by + ?
		WHERE id = ?`,
		[]uuid.UUID{userID},
		messageID,
	).WithContext(ctx).Exec()
}

func (r *messageRepository) MarkAsDelivered(ctx context.Context, messageID uuid.UUID, userID uuid.UUID) error {
	return r.session.Query(`
		UPDATE messages
		SET delivered_to = delivered_to + ?
		WHERE id = ?`,
		[]uuid.UUID{userID},
		messageID,
	).WithContext(ctx).Exec()
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	message.IsEdited = true
	message.EditTimestamp = &time.Time{}
	*message.EditTimestamp = time.Now()

	return r.session.Query(`
		UPDATE messages
		SET content = ?,
		    content_type = ?,
		    is_edited = ?,
		    edit_timestamp = ?
		WHERE id = ?`,
		message.Content,
		message.ContentType,
		message.IsEdited,
		message.EditTimestamp,
		message.ID,
	).WithContext(ctx).Exec()
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.session.Query(`
		DELETE FROM messages
		WHERE id = ?`,
		id,
	).WithContext(ctx).Exec()
}
