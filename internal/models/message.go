package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID            string     `json:"id" bson:"id"`
	SenderID      string     `json:"sender_id" bson:"sender_id"`
	RecipientID   *string    `json:"recipient_id,omitempty" bson:"recipient_id,omitempty"`
	GroupID       *string    `json:"group_id,omitempty" bson:"group_id,omitempty"`
	Content       string     `json:"content" bson:"content"`
	ContentType   string     `json:"content_type" bson:"content_type"` // "text", "image", etc.
	Timestamp     time.Time  `json:"timestamp" bson:"timestamp"`
	ReadBy        []string   `json:"read_by" bson:"read_by"`
	DeliveredTo   []string   `json:"delivered_to" bson:"delivered_to"`
	ReplyToID     *string    `json:"reply_to_id,omitempty" bson:"reply_to_id,omitempty"`
	Attachments   []string   `json:"attachments,omitempty" bson:"attachments,omitempty"`
	IsEdited      bool       `json:"is_edited" bson:"is_edited"`
	EditTimestamp *time.Time `json:"edit_timestamp,omitempty" bson:"edit_timestamp,omitempty"`
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
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		ReadBy:      make([]string, 0),
		DeliveredTo: make([]string, 0),
		IsEdited:    false,
	}
}
