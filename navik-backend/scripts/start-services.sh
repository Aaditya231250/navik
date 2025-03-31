#!/bin/bash
set -e

# Function to determine which docker compose command to use
get_compose_cmd() {
  if command -v docker-compose &> /dev/null; then
    echo "docker-compose"
  else
    echo "docker compose"
  fi
}

COMPOSE_CMD=$(get_compose_cmd)
echo "Using compose command: $COMPOSE_CMD"

echo "Starting Kafka infrastructure..."
cd deploy/kafka
$COMPOSE_CMD up -d
cd ../../

echo "Waiting for Kafka to stablise..."
sleep 30


echo "Starting Location Service..."
cd cmd/location-service
$COMPOSE_CMD up -d
cd ../../

# Start Nginx Gateway
echo "Starting Nginx Gateway..."
cd deploy/nginx
$COMPOSE_CMD up -d
cd ../../

echo "All services started successfully!"
echo "Access the system at:"
echo "- API Gateway: http://localhost"
echo "- Kafka UI: http://localhost/kafka-ui"
echo "- Location Service API: http://localhost/api/location"