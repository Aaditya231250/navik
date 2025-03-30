#!/bin/bash

# Text colors for better output readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default options
REMOVE_VOLUMES=false
FORCE_MODE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --remove-volumes)
      REMOVE_VOLUMES=true
      shift
      ;;
    --force)
      FORCE_MODE=true
      shift
      ;;
    --help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  --remove-volumes    Remove Kafka data volumes (default: preserve volumes)"
      echo "  --force             Force removal without confirmation"
      echo "  --help              Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Function for graceful error handling
handle_error() {
  echo -e "${RED}Error: $1${NC}"
  exit 1
}

# Function to check if a stack exists
stack_exists() {
  docker stack ls --format "{{.Name}}" | grep -q "^$1$"
  return $?
}

# Function to safely remove a stack with verification
remove_stack() {
  local stack_name=$1
  if stack_exists $stack_name; then
    echo -e "${YELLOW}Removing $stack_name stack...${NC}"
    docker stack rm $stack_name || handle_error "Failed to remove $stack_name stack"
    
    # Wait for stack to be fully removed
    local max_wait=30
    local waited=0
    while stack_exists $stack_name && [ $waited -lt $max_wait ]; do
      echo -e "${YELLOW}Waiting for $stack_name stack to be fully removed...${NC}"
      sleep 2
      waited=$((waited+2))
    done
    
    if stack_exists $stack_name; then
      echo -e "${YELLOW}Warning: $stack_name stack still appears to exist after $max_wait seconds${NC}"
    else
      echo -e "${GREEN}$stack_name stack removed successfully${NC}"
    fi
  else
    echo -e "${GREEN}$stack_name stack doesn't exist, skipping${NC}"
  fi
}

# Confirmation if not in force mode
if [ "$FORCE_MODE" != "true" ]; then
  echo -e "${YELLOW}This will stop all Kafka services and stacks.${NC}"
  if [ "$REMOVE_VOLUMES" == "true" ]; then
    echo -e "${RED}WARNING: All Kafka data volumes will be removed. This cannot be undone.${NC}"
  fi
  
  read -p "Are you sure you want to proceed? (y/n): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${GREEN}Operation canceled.${NC}"
    exit 0
  fi
fi

# Start the shutdown process
echo -e "${GREEN}=== Starting graceful shutdown ====${NC}"

# 1. Stop application services first
echo -e "${YELLOW}Step 1/5: Stopping application services...${NC}"
remove_stack "kafka-apps"

# Add a small delay to allow connections to drain
sleep $(( RANDOM % 3 + 2 ))

# 2. Stop monitoring services
echo -e "${YELLOW}Step 2/5: Stopping monitoring services...${NC}"
remove_stack "kafka-monitoring"

# 3. Stop Kafka management tools
echo -e "${YELLOW}Step 3/5: Stopping Kafka management tools...${NC}"
remove_stack "kafka-tools"

# Wait for management tools to fully stop before stopping brokers
sleep 5

# 4. Stop Kafka brokers in reverse order
echo -e "${YELLOW}Step 4/5: Stopping Kafka brokers...${NC}"
remove_stack "kafka-brokers"

# Wait for all services to stop
echo -e "${YELLOW}Waiting for all services to stop completely...${NC}"
sleep 10

# 5. Remove volumes if requested
if [ "$REMOVE_VOLUMES" == "true" ]; then
  echo -e "${YELLOW}Step 5/5: Removing Kafka data volumes...${NC}"
  
  # Remove base infrastructure (which includes network definition)
  remove_stack "kafka-base"
  
  for volume in kafka-mumbai-data kafka-pune-data kafka-delhi-data; do
    if docker volume ls --format "{{.Name}}" | grep -q "^$volume$"; then
      echo -e "${YELLOW}Removing volume $volume...${NC}"
      docker volume rm $volume || echo -e "${YELLOW}Warning: Failed to remove volume $volume${NC}"
    else
      echo -e "${GREEN}Volume $volume doesn't exist, skipping${NC}"
    fi
  done
  
  # Remove the network
  if docker network ls --format "{{.Name}}" | grep -q "^kafka-network$"; then
    echo -e "${YELLOW}Removing kafka-network...${NC}"
    docker network rm kafka-network || echo -e "${YELLOW}Warning: Failed to remove kafka-network${NC}"
  else
    echo -e "${GREEN}Network kafka-network doesn't exist, skipping${NC}"
  fi
else
  echo -e "${GREEN}Step 5/5: Preserving Kafka data volumes for future use${NC}"
fi

# Check for any remaining stacks and services
echo -e "${YELLOW}Checking for any remaining Kafka-related services...${NC}"
REMAINING_SERVICES=$(docker service ls --filter "name=kafka" --format "{{.Name}}" 2>/dev/null)
if [ -n "$REMAINING_SERVICES" ]; then
  echo -e "${YELLOW}Warning: Some Kafka services may still be running:${NC}"
  echo "$REMAINING_SERVICES"
  echo -e "${YELLOW}You may need to remove them manually.${NC}"
else
  echo -e "${GREEN}No Kafka services remaining.${NC}"
fi

# Final cleanup of dangling containers (optional)
if [ "$REMOVE_VOLUMES" == "true" ]; then
  echo -e "${YELLOW}Cleaning up any dangling containers...${NC}"
  docker container prune -f --filter "label=com.docker.stack.namespace=kafka" >/dev/null 2>&1
fi

# Print summary
echo -e "${GREEN}=== Shutdown Summary ====${NC}"
echo -e "${GREEN}All Kafka stacks have been removed${NC}"
if [ "$REMOVE_VOLUMES" == "true" ]; then
  echo -e "${GREEN}Data volumes have been removed${NC}"
else
  echo -e "${GREEN}Data volumes have been preserved for future deployments${NC}"
fi

echo -e "${GREEN}=== Shutdown completed successfully ====${NC}"
