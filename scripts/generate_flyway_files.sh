#!/usr/bin/env bash

# Configuration
SOURCE_DIR="./docker/postgres/schema"
TARGET_DIR="./docker/flyway"
VERSION_PREFIX="V1"
COUNTER=1

# Create the flyway directory
mkdir -p "$TARGET_DIR"

# Remove existing .sql files (since we'll be using actual files, not symbolic links!)
find "$TARGET_DIR" -name "*.sql" -type f -delete

# Loop through sorted SQL files
for file in $(ls "$SOURCE_DIR"/*.sql | sort); do
  # Extract only the filename (e.g. 02-init-voteindexer.sql)
  filename=$(basename "$file")

  # Convert to Flyway naming convention: replace spaces and hyphens with underscores
  description=$(echo "$filename" | sed -E 's/^[0-9]+-init-//' | sed 's/[-]/_/g' | sed 's/\.sql$//')

  # Format counter to two digits (e.g. 01, 02...)
  counter_padded=$(printf "%02d" $COUNTER)

  # Target filename
  target="$TARGET_DIR/${VERSION_PREFIX}_${counter_padded}__init_${description}.sql"

  # Copy the actual file
  cp "$file" "$target"

  echo "Copied: $target"
  ((COUNTER++))
done

# Show directory tree of target
tree "$TARGET_DIR"
