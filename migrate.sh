#!/bin/bash

# Migration script for SQLite database
# Usage: ./migrate.sh [database_path]
# Example: ./migrate.sh ./jobs.sqlite

set -e

# Default database path
DB_PATH="${1:-./db.sqlite}"
MIGRATIONS_DIR="./database/migrations"

# Check if sqlite3 is installed
if ! command -v sqlite3 &> /dev/null; then
    echo "Error: sqlite3 is not installed. Please install it first."
    exit 1
fi

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "Error: Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

echo "Running migrations on database: $DB_PATH"

# Get all migration files sorted by name
MIGRATIONS=$(find "$MIGRATIONS_DIR" -name "*.sql" | sort)

if [ -z "$MIGRATIONS" ]; then
    echo "No migration files found in $MIGRATIONS_DIR"
    exit 1
fi

# Run each migration
for migration in $MIGRATIONS; do
    echo "Running migration: $(basename $migration)"
    sqlite3 "$DB_PATH" < "$migration"
    if [ $? -eq 0 ]; then
        echo "✓ Successfully applied $(basename $migration)"
    else
        echo "✗ Failed to apply $(basename $migration)"
        exit 1
    fi
done

echo ""
echo "All migrations completed successfully!"

