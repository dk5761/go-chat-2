#!/bin/bash

set -e

# Default values
POSTGRES_DSN="postgres://chatapp:chatapp123@localhost:5432/chat?sslmode=disable"
CASSANDRA_HOSTS="localhost"
CASSANDRA_KEYSPACE="chat"

# Function to display usage
show_usage() {
    echo "Usage: $0 [options] [command]"
    echo "Options:"
    echo "  -p, --postgres-dsn DSN    PostgreSQL connection string"
    echo "  -c, --cassandra-hosts     Cassandra hosts (comma-separated)"
    echo "  -k, --keyspace            Cassandra keyspace"
    echo "Commands:"
    echo "  up                        Apply all migrations"
    echo "  down                      Rollback all migrations"
    echo "  create NAME               Create new migration files"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--postgres-dsn)
            POSTGRES_DSN="$2"
            shift 2
            ;;
        -c|--cassandra-hosts)
            CASSANDRA_HOSTS="$2"
            shift 2
            ;;
        -k|--keyspace)
            CASSANDRA_KEYSPACE="$2"
            shift 2
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# Function to run PostgreSQL migrations
run_postgres_migrations() {
    local cmd=$1
    echo "Running PostgreSQL migrations: $cmd"
    migrate -database "$POSTGRES_DSN" -path ./migrations/postgres "$cmd"
}

# Function to run Cassandra migrations
run_cassandra_migrations() {
    local cmd=$1
    echo "Running Cassandra migrations: $cmd"
    for file in ./migrations/cassandra/*_*.up.cql; do
        if [ -f "$file" ]; then
            echo "Applying: $file"
            cqlsh "$CASSANDRA_HOSTS" -f "$file"
        fi
    done
}

# Function to create new migration files
create_migration() {
    local name=$1
    local timestamp=$(date +%Y%m%d%H%M%S)
    
    # Create PostgreSQL migration files
    migrate create -ext sql -dir ./migrations/postgres -seq "${name}"
    
    # Create Cassandra migration files
    local cql_file="./migrations/cassandra/${timestamp}_${name}.up.cql"
    touch "$cql_file"
    echo "Created Cassandra migration: $cql_file"
}

case "$COMMAND" in
    up)
        run_postgres_migrations "up"
        run_cassandra_migrations "up"
        ;;
    down)
        run_postgres_migrations "down"
        # Note: Cassandra down migrations are not supported
        ;;
    create)
        if [ -z "$2" ]; then
            echo "Error: Migration name required"
            show_usage
            exit 1
        fi
        create_migration "$2"
        ;;
    *)
        show_usage
        exit 1
        ;;
esac 