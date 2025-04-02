#!/bin/bash

if [[ ! -z "$KAFKA_CREATE_TOPICS" ]]; then
  echo "Creating topics: $KAFKA_CREATE_TOPICS"
  IFS=',' read -ra TOPICS <<< "$KAFKA_CREATE_TOPICS"
  for TOPIC in "${TOPICS[@]}"; do
    IFS=':' read -ra PARAMS <<< "$TOPIC"
    TOPIC_NAME=${PARAMS[0]}
    PARTITIONS=${PARAMS[1]}
    REPLICATION=${PARAMS[2]}

    kafka-topics --create \
      --bootstrap-server localhost:29092 \
      --topic "$TOPIC_NAME" \
      --partitions "$PARTITIONS" \
      --replication-factor "$REPLICATION" \
      --if-not-exists
  done
else
  echo "No topics specified for this broker"
fi
