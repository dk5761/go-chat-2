#!/bin/bash

# Make script exit on first error
set -e

echo "Creating development environment directories..."

# Create base directories
mkdir -p .cache/go-mod
mkdir -p .data

# Create specific data directories
for dir in postgres cassandra redis rabbitmq pgadmin; do
    mkdir -p ".data/$dir"
    echo "Created .data/$dir"
done

# Set proper permissions
echo "Setting permissions..."
chmod -R 777 .data
chmod -R 777 .cache

# Ensure directories exist and are accessible
for dir in .data/*; do
    if [ -d "$dir" ]; then
        echo "Verifying directory: $dir"
        touch "$dir/.verify" && rm "$dir/.verify" || {
            echo "Error: Directory $dir is not writable"
            exit 1
        }
    fi
done

echo "✅ Development environment directories created successfully!"
echo "You can now run: docker compose -f docker-compose.dev.yml up -d" 