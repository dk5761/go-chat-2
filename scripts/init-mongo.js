// Create the chat database
db = db.getSiblingDB("chat");

// Create messages collection with schema validation
db.createCollection("messages", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["id", "sender_id", "content", "content_type", "timestamp"],
      properties: {
        id: {
          bsonType: "string",
          description: "must be a string and is required",
        },
        sender_id: {
          bsonType: "string",
          description: "must be a string and is required",
        },
        recipient_id: {
          bsonType: ["string", "null"],
          description: "must be a string or null",
        },
        group_id: {
          bsonType: ["string", "null"],
          description: "must be a string or null",
        },
        content: {
          bsonType: "string",
          description: "must be a string and is required",
        },
        content_type: {
          bsonType: "string",
          enum: ["text", "image", "file"],
          description: "must be one of: text, image, file",
        },
        timestamp: {
          bsonType: "date",
          description: "must be a date and is required",
        },
        read_by: {
          bsonType: "array",
          items: {
            bsonType: "string",
          },
          description: "must be an array of strings",
        },
        delivered_to: {
          bsonType: "array",
          items: {
            bsonType: "string",
          },
          description: "must be an array of strings",
        },
        reply_to_id: {
          bsonType: ["string", "null"],
          description: "must be a string or null",
        },
        attachments: {
          bsonType: "array",
          items: {
            bsonType: "string",
          },
          description: "must be an array of strings",
        },
        is_edited: {
          bsonType: "bool",
          description: "must be a boolean",
        },
        edit_timestamp: {
          bsonType: ["date", "null"],
          description: "must be a date or null",
        },
      },
    },
  },
});

// Create indexes for efficient querying
db.messages.createIndex({ sender_id: 1 });
db.messages.createIndex({ recipient_id: 1 });
db.messages.createIndex({ group_id: 1 });
db.messages.createIndex({ timestamp: -1 });
db.messages.createIndex({ recipient_id: 1, timestamp: -1 });
db.messages.createIndex({ group_id: 1, timestamp: -1 });

// Create a compound index for conversation queries
db.messages.createIndex({
  sender_id: 1,
  recipient_id: 1,
  timestamp: -1,
});

// Create a TTL index for temporary data (e.g., message drafts)
db.messages.createIndex({ timestamp: 1 }, { expireAfterSeconds: 604800 }); // 7 days
