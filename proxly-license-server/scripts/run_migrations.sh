#!/bin/bash

# Migration script for Proxly License Server
# Runs database migrations using golang-migrate

set -e

DB_DSN="${DB_DSN:-postgres://proxly:proxly@localhost:5432/proxly?sslmode=disable}"
MIGRATIONS_PATH="${MIGRATIONS_PATH:-internal/store/migrations}"

echo "🔄 Running database migrations..."
echo "Database: $DB_DSN"
echo "Migrations path: $MIGRATIONS_PATH"

# Check if migrate command is available
if ! command -v migrate &> /dev/null; then
    echo "❌ golang-migrate not found. Installing..."

    # Install golang-migrate
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

    # Add GOPATH/bin to PATH if not already there
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Run migrations
migrate -path "$MIGRATIONS_PATH" -database "$DB_DSN" up

echo "✅ Database migrations completed successfully!"