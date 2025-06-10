#!/bin/sh
set -e

# --- Source Env Vars ---
if [ -f dotenv ]; then
  echo "INFO: Sourcing environment variables from dotenv file..."
  . ./dotenv
else
  echo "ERROR: dotenv file not found. Cannot determine image names." >&2
  exit 1
fi

# Check if variables are set
: "${IMAGE_NAME_APPLY:?ERROR: IMAGE_NAME_APPLY not set in dotenv}"
: "${IMAGE_NAME_DESTROY:?ERROR: IMAGE_NAME_DESTROY not set in dotenv}"

# Ensure user is logged into Docker Hub (or target registry)
echo "INFO: Please ensure you are logged into your container registry (e.g., docker login)"

echo "Pushing LEMC Terraform Apply image: ${IMAGE_NAME_APPLY}"
docker push "${IMAGE_NAME_APPLY}"

echo "Pushing LEMC Terraform Destroy image: ${IMAGE_NAME_DESTROY}"
docker push "${IMAGE_NAME_DESTROY}"

echo "Push complete." 