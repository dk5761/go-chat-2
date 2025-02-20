package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Message struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SenderID      uuid.UUID      `json:"sender_id" gorm:"type:uuid;not null"`
	RecipientID   *uuid.UUID     `json:"recipient_id,omitempty" gorm:"type:uuid"`
	GroupID       *uuid.UUID     `json:"group_id,omitempty" gorm:"type:uuid"`
	Content       string         `json:"content" gorm:"not null"`
	ContentType   string         `json:"content_type" gorm:"not null"` // "text", "image", etc.
	Timestamp     time.Time      `json:"timestamp" gorm:"not null;default:CURRENT_TIMESTAMP"`
	ReadBy        pq.StringArray `json:"read_by" gorm:"type:text[]"`
	DeliveredTo   pq.StringArray `json:"delivered_to" gorm:"type:text[]"`
	ReplyToID     *uuid.UUID     `json:"reply_to_id,omitempty" gorm:"type:uuid"`
	Attachments   pq.StringArray `json:"attachments,omitempty" gorm:"type:text[]"`
	IsEdited      bool           `json:"is_edited" gorm:"not null;default:false"`
	EditTimestamp *time.Time     `json:"edit_timestamp,omitempty"`
}

// CQL table creation statement
const MessageTableCQL = `
CREATE TABLE IF NOT EXISTS messages (
    id uuid,
    sender_id uuid,
    recipient_id uuid,
    group_id uuid,
    content text,
    content_type text,
    timestamp timestamp,
    read_by list<uuid>,
    delivered_to list<uuid>,
    reply_to_id uuid,
    attachments list<text>,
    is_edited boolean,
    edit_timestamp timestamp,
    PRIMARY KEY ((recipient_id, group_id), timestamp, id)
) WITH CLUSTERING ORDER BY (timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_group ON messages (group_id);
`

// Message status constants
const (
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
	MessageStatusFailed    = "failed"
)

// Content type constants
const (
	ContentTypeText  = "text"
	ContentTypeImage = "image"
	ContentTypeFile  = "file"
)

// NewMessage creates a new message with a generated UUID and current timestamp
func NewMessage() *Message {
	return &Message{
		ID:          uuid.New(),
		Timestamp:   time.Now(),
		ReadBy:      make(pq.StringArray, 0),
		DeliveredTo: make(pq.StringArray, 0),
		IsEdited:    false,
	}
}
