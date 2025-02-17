CREATE TABLE IF NOT EXISTS chat.messages (
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

CREATE INDEX IF NOT EXISTS idx_messages_sender ON chat.messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_group ON chat.messages (group_id); 