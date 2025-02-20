# Chat Application API Endpoints

## Public Routes (No Authentication Required)

### Health Check
- GET /api/v1/health - Check API health status

### User Management
- POST /api/v1/users/register - Register new user
- POST /api/v1/users/login - User login

## Protected Routes (Authentication Required)

### User Operations
- GET /api/v1/users/:id - Get user details
- PUT /api/v1/users/:id/password - Update user password
- GET /api/v1/users/:id/status - Get user's online status
- POST /api/v1/users/status/multi - Get multiple users' statuses

### Group Operations
- POST /api/v1/groups - Create new group
- GET /api/v1/groups/:id - Get group details
- PUT /api/v1/groups/:id - Update group details
- DELETE /api/v1/groups/:id - Delete group
- POST /api/v1/groups/:id/members - Add member to group
- DELETE /api/v1/groups/:id/members/:user_id - Remove member from group
- GET /api/v1/groups/:id/members - Get all group members
- GET /api/v1/groups/user/:id - Get user's groups
- PUT /api/v1/groups/:id/members/:user_id/role - Update member's role

### Message Operations
- POST /api/v1/messages - Send new message
- GET /api/v1/messages/:id - Get message by ID
- GET /api/v1/messages/user/:id - Get user's messages
- GET /api/v1/messages/group/:id - Get group messages
- GET /api/v1/messages/conversation/:user1_id/:user2_id - Get conversation between two users
- POST /api/v1/messages/:id/read - Mark message as read
- DELETE /api/v1/messages/:id - Delete message

### WebSocket
- GET /api/v1/ws - WebSocket connection endpoint

## WebSocket Message Types

### Direct Messages
```json
{
  "type": "chat",
  "sender_id": "uuid",
  "recipient_id": "uuid",
  "content": "message content",
  "content_type": "text|image",
  "timestamp": "ISO8601"
}
```

### Group Messages
```json
{
  "type": "chat",
  "sender_id": "uuid",
  "group_id": "uuid",
  "content": "message content",
  "content_type": "text|image",
  "timestamp": "ISO8601"
}
```

### Typing Indicators
```json
{
  "type": "typing",
  "sender_id": "uuid",
  "recipient_id": "uuid",
  "content": "typing",
  "timestamp": "ISO8601"
}
```

### Read Receipts
```json
{
  "type": "read",
  "sender_id": "uuid",
  "recipient_id": "uuid",
  "content": "message_uuid",
  "timestamp": "ISO8601"
}
```

## Authentication
- All protected routes require Bearer token authentication
- Token format: `Bearer <jwt_token>`
- Token obtained from login response

## Query Parameters
- Messages endpoints support:
  - `limit` (default: 50) - Number of messages to return
  - `offset` (default: 0) - Offset for pagination

## Response Formats
- Success responses: HTTP 2xx with JSON body
- Error responses: HTTP 4xx/5xx with JSON error message 