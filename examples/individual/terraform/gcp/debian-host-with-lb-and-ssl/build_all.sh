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

echo "Building LEMC Terraform Apply image: ${IMAGE_NAME_APPLY}"
docker build -t "${IMAGE_NAME_APPLY}" -f ./lemc-tf-apply/Dockerfile .

echo "Building LEMC Terraform Destroy image: ${IMAGE_NAME_DESTROY}"
docker build -t "${IMAGE_NAME_DESTROY}" -f ./lemc-tf-destroy/Dockerfile .

echo "Build complete."