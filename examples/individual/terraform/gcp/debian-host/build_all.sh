#!/bin/bash
set -e # Exit on error

TF_CONFIG_DIR="./terraform-config"

# --- Pre-build Checks ---
if [ ! -d "$TF_CONFIG_DIR" ]; then
  echo "Error: Terraform configuration directory not found: $TF_CONFIG_DIR" >&2
  echo "Please create this directory and place your .tf files inside it." >&2
  exit 1
fi

if ! ls "${TF_CONFIG_DIR}"/*.tf > /dev/null 2>&1; then
  echo "Error: No Terraform configuration files (.tf) found in $TF_CONFIG_DIR" >&2
  exit 1
fi

echo "Found Terraform configuration in $TF_CONFIG_DIR"

# Build lemc-tf-apply
echo "Building lemc-tf-apply..."
# Use '.' as context, specify Dockerfile with -f
docker build -t docker.io/jfolkins/lemc-tf-apply:latest -f lemc-tf-apply/Dockerfile .

# Build lemc-tf-destroy
echo "Building lemc-tf-destroy..."
# Use '.' as context, specify Dockerfile with -f
docker build -t docker.io/jfolkins/lemc-tf-destroy:latest -f lemc-tf-destroy/Dockerfile .

echo "Build complete."