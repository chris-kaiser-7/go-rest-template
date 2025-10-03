#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="test-kafka"
PORT="9092:9092"

# Stop and remove any existing container with the same name
if docker ps -a --format '{{.Names}}' | grep -Eq "^${CONTAINER_NAME}$"; then
    echo "Stopping and removing existing container: ${CONTAINER_NAME}"
    docker stop "$CONTAINER_NAME"
    docker rm "$CONTAINER_NAME"
fi

echo "Starting new Kafka container..."
docker run -d \
  --name "$CONTAINER_NAME" \
  -p "$PORT" \
  apache/kafka:4.1.0

docker exec $CONTAINER_NAME /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --topic test

echo "$CONTAINER_NAME is running on port: $PORT"
