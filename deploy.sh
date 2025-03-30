#!/bin/bash

# Text colors for better output readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to wait for Kafka broker to be healthy with exponential backoff
wait_for_kafka() {
  local broker=$1
  local port=$2
  local max_attempts=10
  local timeout=5
  local attempt=1
  local backoff=2
  
  echo -e "${YELLOW}Waiting for Kafka broker $broker to be ready...${NC}"
  
  while [ $attempt -le $max_attempts ]; do
    echo -e "Attempt $attempt/$max_attempts: Checking Kafka broker $broker"
    
    if docker run --rm --network kafka-network confluentinc/cp-kafka:7.5.1 kafka-topics --bootstrap-server $broker:$port --list > /dev/null 2>&1; then
      echo -e "${GREEN}Kafka broker $broker is ready!${NC}"
      return 0
    fi
    
    echo -e "${YELLOW}Kafka broker $broker not yet ready. Retrying in $timeout seconds...${NC}"
    sleep $timeout
    
    # Increase timeout with exponential backoff (with a cap)
    timeout=$(( timeout * backoff ))
    if [ $timeout -gt 60 ]; then
      timeout=60
    fi
    
    attempt=$(( attempt + 1 ))
  done
  
  echo -e "${RED}Failed to connect to Kafka broker $broker after $max_attempts attempts.${NC}"
  return 1
}

# Function for graceful error handling
handle_error() {
  echo -e "${RED}Error: $1${NC}"
  exit 1
}

# Check if Docker Swarm is initialized
if ! docker info | grep -q "Swarm: active"; then
  echo -e "${YELLOW}Docker Swarm not active. Initializing Swarm...${NC}"
  docker swarm init --advertise-addr $(hostname -I | awk '{print $1}') || handle_error "Failed to initialize Docker Swarm"
fi

# Create external network if it doesn't exist
if ! docker network ls | grep -q "kafka-network"; then
  echo -e "${YELLOW}Creating Kafka network...${NC}"
  docker network create --driver overlay --attachable kafka-network || handle_error "Failed to create network"
  echo -e "${GREEN}Kafka network created successfully${NC}"
else
  echo -e "${GREEN}Kafka network already exists${NC}"
fi

# Create external volumes
echo -e "${YELLOW}Creating Kafka volumes...${NC}"
for volume in kafka-mumbai-data kafka-pune-data kafka-delhi-data; do
  if ! docker volume ls | grep -q "$volume"; then
    docker volume create $volume || handle_error "Failed to create volume $volume"
    echo -e "${GREEN}Volume $volume created successfully${NC}"
  else
    echo -e "${GREEN}Volume $volume already exists${NC}"
  fi
done

# Deploy base infrastructure
echo -e "${YELLOW}Deploying base infrastructure...${NC}"
docker stack deploy -c base.yml kafka-base || handle_error "Failed to deploy base infrastructure"
echo -e "${GREEN}Base infrastructure deployed successfully${NC}"

# Deploy Kafka brokers with proper sequence and health checks
echo -e "${YELLOW}Deploying Kafka brokers...${NC}"
docker stack deploy -c kafka-brokers.yml kafka-brokers || handle_error "Failed to deploy Kafka brokers"

# Wait for first broker (Mumbai) to be ready before proceeding
echo -e "${YELLOW}Waiting for primary Kafka broker to initialize...${NC}"
sleep 10  # Initial sleep to allow container to start
wait_for_kafka "kafka-mumbai" "29092" || handle_error "Primary Kafka broker failed to start"

# Wait for the other brokers with jittered sleep
echo -e "${YELLOW}Waiting for additional Kafka brokers to initialize...${NC}"
# Add jitter to prevent connection storms
sleep $(( RANDOM % 5 + 5 ))
wait_for_kafka "kafka-pune" "29092" || echo -e "${YELLOW}Warning: Pune broker may not be ready, but continuing...${NC}"
sleep $(( RANDOM % 5 + 3 ))
wait_for_kafka "kafka-delhi" "29092" || echo -e "${YELLOW}Warning: Delhi broker may not be ready, but continuing...${NC}"

# Check overall Kafka cluster health
echo -e "${YELLOW}Verifying Kafka cluster health...${NC}"
if docker run --rm --network kafka-network confluentinc/cp-kafka:7.5.1 kafka-broker-api-versions --bootstrap-server kafka-mumbai:29092,kafka-pune:29092,kafka-delhi:29092 > /dev/null 2>&1; then
  echo -e "${GREEN}Kafka cluster is healthy!${NC}"
else
  echo -e "${YELLOW}Warning: Kafka cluster may not be fully healthy yet, but proceeding...${NC}"
fi

# Deploy Kafka management tools with jittered sleep to prevent connection storms
echo -e "${YELLOW}Deploying Kafka management tools...${NC}"
sleep $(( RANDOM % 3 + 2 ))
docker stack deploy -c kafka-tools.yml kafka-tools || handle_error "Failed to deploy Kafka tools"

# Wait for topics to be created with progressive notification
echo -e "${YELLOW}Initializing Kafka topics...${NC}"
topic_init_wait=45
progress_interval=5
for (( i=0; i<$topic_init_wait; i+=$progress_interval )); do
  sleep $progress_interval
  echo -e "${YELLOW}Topic initialization in progress... $(( (i+$progress_interval) * 100 / $topic_init_wait ))% complete${NC}"
done

# Verify topics were created
echo -e "${YELLOW}Verifying topic creation...${NC}"
if docker run --rm --network kafka-network confluentinc/cp-kafka:7.5.1 kafka-topics --bootstrap-server kafka-mumbai:29092 --list | grep -q "locations"; then
  echo -e "${GREEN}Kafka topics created successfully${NC}"
else
  echo -e "${YELLOW}Warning: Topics may not be fully created yet, but proceeding...${NC}"
fi

# Deploy application services with staggered starts to avoid connection storms
echo -e "${YELLOW}Deploying application services...${NC}"
sleep $(( RANDOM % 5 + 3 ))
docker stack deploy -c application.yml kafka-apps || handle_error "Failed to deploy application services"

# Deploy monitoring with a slight delay
echo -e "${YELLOW}Deploying monitoring...${NC}"
sleep $(( RANDOM % 3 + 2 ))
docker stack deploy -c monitoring.yml kafka-monitoring || handle_error "Failed to deploy monitoring"

# Print success message with stack information
echo -e "${GREEN}Deployment complete!${NC}"
echo -e "${GREEN}=== Deployment Summary ===${NC}"
echo -e "${GREEN}Kafka UI:${NC} http://localhost:8080"
echo -e "${GREEN}Cluster Visualizer:${NC} http://localhost:8081"
echo -e "${GREEN}Producer API:${NC} http://localhost:6969"

# Print deployment status
echo -e "${YELLOW}Current stack status:${NC}"
docker stack ls
echo -e "${YELLOW}Service status:${NC}"
docker service ls

echo -e "${GREEN}=== Deployment completed successfully ===${NC}"
echo -e "${YELLOW}Note: It may take a few more minutes for all services to be fully operational${NC}"
