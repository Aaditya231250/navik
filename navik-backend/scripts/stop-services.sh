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

# Stop Nginx Gateway
echo "Stopping Nginx Gateway..."
cd deploy/nginx
$COMPOSE_CMD down
cd ../../

# Stop Location Service
echo "Stopping Location Service..."
cd cmd/location-service
$COMPOSE_CMD down
cd ../../

# Stop DynamoDB infrastructure
echo "Stopping DynamoDB infrastructure..."
cd deploy/dynamodb
$COMPOSE_CMD down
cd ../../

# Stop Kafka infrastructure
echo "Stopping Kafka infrastructure..."
cd deploy/kafka
$COMPOSE_CMD down
cd ../../

echo "All services stopped successfully!"
