#!/bin/bash

set -euo pipefail

echo "Waiting for Kafka to be available at kafka:29092..."
until kafka-topics --bootstrap-server kafka:29092 --list >/dev/null 2>&1; do
  echo "Kafka not available yet. Sleeping 1s..."
  sleep 1
done

echo -e 'Preparing to create kafka topics'

# TOPICS is expected to be set like: TOPICS="topic1,topic2,topic3
topics_str="${TOPICS//,/ }"
read -r -a topics <<< "$topics_str"

if [ "${#topics[@]}" -eq 0 ]; then
  echo "No topics found in TOPICS environment variable. Nothing to create."
  exit 0
fi

echo "Creating topics: ${topics[*]}"

for topic in "${topics[@]}"; do
  # Trim leading and trailing whitespace using bash parameter expansion
  topic="${topic#"${topic%%[![:space:]]*}"}"  # remove leading
  topic="${topic%"${topic##*[![:space:]]}"}"  # remove trailing

  if [ -z "$topic" ]; then
    continue
  fi

  echo "Creating topic: $topic"
  kafka-topics --bootstrap-server kafka:29092 --create --if-not-exists --topic "$topic"
done

echo -e 'Successfully created the following topics:'
kafka-topics --bootstrap-server kafka:29092 --list
