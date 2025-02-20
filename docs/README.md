# Chat Application Documentation

This directory contains documentation for the Chat Application.

## Contents

- [API Documentation](api.md) - Complete API reference including endpoints, WebSocket messages, and authentication
- [Database Schema](schema.md) - Database schema and relationships
- [Architecture](architecture.md) - System architecture and component interactions

## API Documentation

The [API documentation](api.md) includes:
- REST API endpoints
- WebSocket message formats
- Authentication details
- Query parameters
- Response formats

## Using the API

1. Start with the public endpoints to register and login
2. Use the JWT token from login for authenticated requests
3. Connect to WebSocket for real-time messaging
4. Use the appropriate message formats for different types of communication

## WebSocket Testing

Use the Postman collections in the `postman` directory for testing:
- `chat-websocket.postman_collection.json` - WebSocket connection tests
- `websocket-messages.json` - Message templates for testing

## Contributing

When adding new API endpoints or modifying existing ones:
1. Update the API documentation
2. Add appropriate test cases
3. Update the Postman collections 