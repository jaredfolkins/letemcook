#!/bin/bash
set -e # Exit on error

# --- Pre-push Checks ---
echo "Checking for Docker images..."

# Check if images exist
if ! docker image inspect docker.io/jfolkins/lemc-tf-apply:latest > /dev/null 2>&1; then
  echo "Error: Image docker.io/jfolkins/lemc-tf-apply:latest not found" >&2
  echo "Please run build_all.sh first to build the images." >&2
  exit 1
fi

if ! docker image inspect docker.io/jfolkins/lemc-tf-destroy:latest > /dev/null 2>&1; then
  echo "Error: Image docker.io/jfolkins/lemc-tf-destroy:latest not found" >&2
  echo "Please run build_all.sh first to build the images." >&2
  exit 1
fi

echo "Images found, proceeding with push..."

# Push lemc-tf-apply
echo "Pushing lemc-tf-apply..."
docker push docker.io/jfolkins/lemc-tf-apply:latest

# Push lemc-tf-destroy
echo "Pushing lemc-tf-destroy..."
docker push docker.io/jfolkins/lemc-tf-destroy:latest

echo "Push complete." 