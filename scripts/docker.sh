#!/bin/bash

# Make script exit on first error
set -e

# Default values
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# Function to display usage
show_usage() {
    echo "Usage: $0 [command]"
    echo "Commands:"
    echo "  up        - Start all services"
    echo "  down      - Stop and remove all services"
    echo "  restart   - Restart all services"
    echo "  logs      - Show logs from all services"
    echo "  ps        - Show status of services"
    echo "  build     - Rebuild services"
    echo "  clean     - Remove all containers, volumes, and images"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo "Error: Docker is not running"
        exit 1
    fi
}

# Main script
check_docker

case "$1" in
    up)
        echo "Starting services..."
        docker-compose -f $COMPOSE_FILE up -d
        echo "Services started. Use '$0 logs' to view logs"
        ;;
    down)
        echo "Stopping services..."
        docker-compose -f $COMPOSE_FILE down
        ;;
    restart)
        echo "Restarting services..."
        docker-compose -f $COMPOSE_FILE restart
        ;;
    logs)
        docker-compose -f $COMPOSE_FILE logs -f
        ;;
    ps)
        docker-compose -f $COMPOSE_FILE ps
        ;;
    build)
        echo "Rebuilding services..."
        docker-compose -f $COMPOSE_FILE build --no-cache
        ;;
    clean)
        echo "Cleaning up Docker resources..."
        docker-compose -f $COMPOSE_FILE down -v --rmi all
        ;;
    *)
        show_usage
        exit 1
        ;;
esac 