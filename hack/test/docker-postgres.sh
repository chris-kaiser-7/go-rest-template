#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="test-postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_USER="postgres"
POSTGRES_DB="testdb"
PORT="5433"

# Stop and remove any existing container with the same name
if docker ps -a --format '{{.Names}}' | grep -Eq "^${CONTAINER_NAME}$"; then
    echo "Stopping and removing existing container: ${CONTAINER_NAME}"
    docker stop "$CONTAINER_NAME"
    docker rm "$CONTAINER_NAME"
fi

# Start a new PostgreSQL container
echo "Starting new Postgres container..."
docker run -d \
  --name "$CONTAINER_NAME" \
  -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  -e POSTGRES_USER="$POSTGRES_USER" \
  -e POSTGRES_DB="$POSTGRES_DB" \
  -p "$PORT:5432" \
  postgres:17


# Wait for PostgreSQL to become ready
echo "Waiting for PostgreSQL to be ready..."
until docker exec "$CONTAINER_NAME" pg_isready -U "$POSTGRES_USER" > /dev/null 2>&1; do
  sleep 1
done

sleep 1

docker exec "test-postgres" psql --username=postgres -d testdb -c "CREATE EXTENSION IF NOT EXISTS citext"

echo "PostgreSQL is ready and running at localhost:$PORT"
