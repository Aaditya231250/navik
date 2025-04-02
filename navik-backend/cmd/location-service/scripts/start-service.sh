#!/bin/bash
set -e

SERVICE_TYPE=$1
echo "Starting $SERVICE_TYPE service with smart retry logic"

# Configuration (from environment or defaults)
MAX_ATTEMPTS=${KAFKA_RETRY_MAX:-5}
BACKOFF_MS=${KAFKA_RETRY_INITIAL_MS:-1000}
MAX_BACKOFF_MS=${KAFKA_RETRY_MAX_MS:-16000}
BROKER_LIST=${KAFKA_BROKERS:-"kafka-mumbai:29092"}
BINARY=${SERVICE_BINARY:-/app/service}

# Extract first broker for checks
BROKER=$(echo $BROKER_LIST | cut -d ',' -f1)
echo "Using broker $BROKER for availability checks"

# Check for available tools
if command -v kafka-topics &> /dev/null; then
    KAFKA_TOOL="kafka-topics"
    echo "Using kafka-topics CLI tool"
elif command -v kafkacat &> /dev/null; then
    KAFKA_TOOL="kafkacat"
    echo "Using kafkacat tool"
else
    echo "WARNING: No Kafka tools found. Installing kafkacat..."
    apk add --no-cache kafkacat &> /dev/null || echo "Failed to install kafkacat"
    KAFKA_TOOL="kafkacat"
fi

# Function to check if Kafka is reachable
check_kafka_connection() {
    echo "Checking Kafka connection to $BROKER..."
    if [ "$KAFKA_TOOL" = "kafka-topics" ]; then
        timeout 5 kafka-topics --bootstrap-server $BROKER --list &> /dev/null
        return $?
    else
        timeout 5 kafkacat -b $BROKER -L -t non-existent-topic -T 3 &> /dev/null
        return $?
    fi
}

# Function to check if Kafka topics exist
check_topics() {
    echo "Checking for 'locations' topics on $BROKER..."
    if [ "$KAFKA_TOOL" = "kafka-topics" ]; then
        topics=$(kafka-topics --bootstrap-server $BROKER --list 2>/dev/null)
        echo "Available topics: $topics"
        echo "$topics" | grep -q 'locations'
        return $?
    else
        topics=$(kafkacat -b $BROKER -L 2>/dev/null | grep -o "topic \"[^\"]*\"" | cut -d '"' -f2)
        echo "Available topics: $topics"
        echo "$topics" | grep -q 'locations'
        return $?
    fi
}

# Main connection logic with exponential backoff
ATTEMPT=1
echo "🔄 Starting connection attempts to Kafka..."

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    SLEEP_SEC=$((BACKOFF_MS/1000))
    
    echo "🔄 Attempt $ATTEMPT/$MAX_ATTEMPTS: Checking Kafka availability..."
    
    # First check if Kafka is reachable
    if check_kafka_connection; then
        echo "✅ Kafka connection successful"
        
        # Then check if topics exist
        if check_topics; then
            echo "✅ Kafka topics found, starting $SERVICE_TYPE service"
            exec $BINARY
            exit 0  # This line won't be reached if exec succeeds
        else
            echo "⚠️ Kafka available but topics not found"
        fi
    else
        echo "⚠️ Kafka connection failed"
    fi
    
    echo "⏳ Waiting ${SLEEP_SEC}s before retry..."
    sleep $SLEEP_SEC
    
    # Exponential backoff with cap
    BACKOFF_MS=$((BACKOFF_MS*2))
    if [ $BACKOFF_MS -gt $MAX_BACKOFF_MS ]; then
        BACKOFF_MS=$MAX_BACKOFF_MS
    fi
    
    ATTEMPT=$((ATTEMPT+1))
done

echo "⚠️ Failed to connect to Kafka or find topics after $MAX_ATTEMPTS attempts"
echo "🔄 Starting $SERVICE_TYPE anyway, hoping Kafka will be available when needed"
exec $BINARY
