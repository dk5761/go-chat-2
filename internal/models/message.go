package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID            uuid.UUID   `json:"id" cql:"id"`
	SenderID      uuid.UUID   `json:"sender_id" cql:"sender_id"`
	RecipientID   *uuid.UUID  `json:"recipient_id,omitempty" cql:"recipient_id"`
	GroupID       *uuid.UUID  `json:"group_id,omitempty" cql:"group_id"`
	Content       string      `json:"content" cql:"content"`
	ContentType   string      `json:"content_type" cql:"content_type"` // "text", "image", etc.
	Timestamp     time.Time   `json:"timestamp" cql:"timestamp"`
	ReadBy        []uuid.UUID `json:"read_by" cql:"read_by"`
	DeliveredTo   []uuid.UUID `json:"delivered_to" cql:"delivered_to"`
	ReplyToID     *uuid.UUID  `json:"reply_to_id,omitempty" cql:"reply_to_id"`
	Attachments   []string    `json:"attachments,omitempty" cql:"attachments"`
	IsEdited      bool        `json:"is_edited" cql:"is_edited"`
	EditTimestamp *time.Time  `json:"edit_timestamp,omitempty" cql:"edit_timestamp"`
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
