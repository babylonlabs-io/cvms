#!/usr/bin/env bash

set -e

# Load .env file (if exists)
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# Default values (fallback if environment variables are not defined)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-mydb}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-secret}"

# Path to Flyway SQL files
FLYWAY_SQL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../docker/flyway" && pwd)"

echo "💡 Running Flyway migrations against $DB_HOST:$DB_PORT/$DB_NAME..."

docker run --rm \
  -v $(pwd)/docker/flyway:/flyway/sql \
  flyway/flyway:latest \
  -url="jdbc:postgresql://$DB_HOST:$DB_PORT/$DB_NAME" \
  -user="$DB_USER" \
  -password="$DB_PASSWORD" \
  -locations=filesystem:/flyway/sql \
  migrate
