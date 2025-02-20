-- Create keyspace
CREATE KEYSPACE IF NOT EXISTS chat
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
};

-- Switch to the keyspace
USE chat;

-- Create messages table
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_group ON messages (group_id); 