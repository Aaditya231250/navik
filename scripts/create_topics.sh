#!/bin/bash

echo "Waiting for Kafka to be ready..."
sleep 10

kafka-topics --create --if-not-exists \
  --topic mumbai-locations \
  --bootstrap-server kafka-mumbai:29092 \
  --partitions 2 \
  --replication-factor 3

kafka-topics --create --if-not-exists \
  --topic pune-locations \
  --bootstrap-server kafka-pune:29092 \
  --partitions 2 \
  --replication-factor 3

kafka-topics --create --if-not-exists \
  --topic delhi-locations \
  --bootstrap-server kafka-delhi:29092 \
  --partitions 2 \
  --replication-factor 3

echo "Topics created successfully!"
