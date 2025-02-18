#!/bin/bash
set -e

echo "Creating development environment directories..."

# Get current user and group
USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Create base directories
mkdir -p .cache/go-mod
mkdir -p .data

# Create specific data directories
mkdir -p .data/postgres
mkdir -p .data/redis
mkdir -p .data/pgadmin
mkdir -p .data/rabbitmq

# Set permissions for PostgreSQL
sudo chown -R $USER_ID:$GROUP_ID .data/postgres
sudo chmod -R 755 .data/postgres

# Set permissions for Redis
sudo chown -R $USER_ID:$GROUP_ID .data/redis
sudo chmod -R 755 .data/redis

# Set permissions for pgAdmin
sudo chown -R $USER_ID:$GROUP_ID .data/pgadmin
sudo chmod -R 755 .data/pgadmin

# Set permissions for RabbitMQ
sudo chown -R $USER_ID:$GROUP_ID .data/rabbitmq
sudo chmod -R 755 .data/rabbitmq

# Set permissions for Go mod cache
sudo chown -R $USER_ID:$GROUP_ID .cache/go-mod
sudo chmod -R 755 .cache/go-mod

echo "✅ Development environment directories created successfully!"
echo "You can now run: docker compose -f docker-compose.dev.yml up -d"
