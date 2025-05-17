#!/bin/sh
# Bootstrap a testing environment and run the test suite
set -e

# Ensure we operate from repo root
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# Configure environment variables for testing
export LEMC_ENV=test
export LEMC_DATA="${REPO_ROOT}/data"
export LEMC_PORT_TEST=${LEMC_PORT_TEST:-15362}

# Create the test data directory if it doesn't exist
mkdir -p "$LEMC_DATA/$LEMC_ENV"

# Start the server in the background
# It will initialize the database and seed with development data
# (SeedDatabaseIfDev only seeds when env is development, so manually invoke when test)

go run main.go &
LEMCPID=$!
trap 'kill $LEMCPID' EXIT

# Give the server a moment to start
sleep 2

go test ./...

# Shutdown the server when tests finish
kill $LEMCPID
