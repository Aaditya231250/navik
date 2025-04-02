#!/bin/bash
set -e

echo "Starting Kafka topic initialization"

# Configuration
MAX_ATTEMPTS=5
ATTEMPT=1
BACKOFF_TIME=5
BROKER="kafka-mumbai:29092"
TOPICS=("mumbai-locations" "pune-locations" "delhi-locations")
PARTITIONS=2
REPLICATION=3

# Function to create topics
create_topics() {
  for topic in "${TOPICS[@]}"; do
    echo "Creating topic: $topic"
    kafka-topics --bootstrap-server $BROKER --create --if-not-exists \
      --topic "$topic" --partitions $PARTITIONS --replication-factor $REPLICATION
    
    if [ $? -eq 0 ]; then
      echo "✅ Topic $topic created successfully"
    else
      echo "❌ Failed to create topic $topic"
      return 1
    fi
  done
  return 0
}

# Main execution with exponential backoff
while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  echo "🔄 Attempt $ATTEMPT of $MAX_ATTEMPTS (waiting ${BACKOFF_TIME}s before trying)"
  sleep $BACKOFF_TIME
  
  # Check if Kafka is ready
  if kafka-topics --bootstrap-server $BROKER --list &>/dev/null; then
    echo "✅ Kafka is ready. Creating topics..."
    
    if create_topics; then
      echo "✅ All topics created successfully"
      # List all topics for verification
      echo "Current topics:"
      kafka-topics --bootstrap-server $BROKER --list
      exit 0
    fi
  else
    echo "⏳ Kafka not yet ready..."
  fi
  
  # Exponential backoff
  ATTEMPT=$((ATTEMPT+1))
  BACKOFF_TIME=$((BACKOFF_TIME*2))
done

echo "❌ Failed to create topics after maximum attempts"
exit 1
