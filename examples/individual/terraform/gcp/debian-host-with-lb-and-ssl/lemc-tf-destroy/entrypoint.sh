#!/bin/sh
set -e # Exit immediately if a command exits with a non-zero status.

echo "--- Entrypoint Start (Destroy) ---"

# --- Configuration & Setup ---
TF_WORKING_DIR="/lemc/private"
CRED_FILE="/tmp/gcp-credentials.json"
IMAGE_TF_CONFIG_DIR="/app/tf_config" # Location of files copied in Dockerfile

mkdir -p "$TF_WORKING_DIR"
cd "$TF_WORKING_DIR" || exit 1

TF_COMMAND="destroy -auto-approve"
ACTION_NAME="Destroy"

echo "Action: Terraform $ACTION_NAME in $TF_WORKING_DIR"

# Copy Terraform files from the image's internal location
if [ -d "$IMAGE_TF_CONFIG_DIR" ] && [ "$(ls -A $IMAGE_TF_CONFIG_DIR/*.tf 2>/dev/null)" ]; then
  echo "Copying TF files from image path $IMAGE_TF_CONFIG_DIR to $TF_WORKING_DIR..."
  cp "$IMAGE_TF_CONFIG_DIR"/*.tf ./
else
  echo "Warning: No .tf files found in image path $IMAGE_TF_CONFIG_DIR."
fi

# Check if TF files exist in the working directory now
if ! ls *.tf > /dev/null 2>&1; then
  echo "lemc.html.append;<p style='color:red;font-weight:bold;'>Error: No Terraform configuration files (.tf) found in working directory $TF_WORKING_DIR after attempting copy from image. Cannot perform destroy.</p>"
  exit 1
fi

# Check for state file (optional but good practice)
if [ ! -f "terraform.tfstate" ]; then
    echo "lemc.html.append;<p style='color:orange;'>Warning: No terraform.tfstate file found in $TF_WORKING_DIR. Destroy may fail or do nothing if resources weren't created in this context.</p>"
fi

# --- Credential Handling ---
# Priority: Base64 Env Var > ADC
# Use direct variable name GCP_SA_KEY_JSON_B64
if [ -n "$GCP_SA_KEY_JSON_B64" ]; then
  echo "Found Base64 encoded GCP credentials in GCP_SA_KEY_JSON_B64."
  echo "Decoding to $CRED_FILE..."
  echo "$GCP_SA_KEY_JSON_B64" | base64 -d > "$CRED_FILE"
  if [ $? -ne 0 ]; then
      echo "lemc.html.append;<p style='color:red;'>Error: Failed to decode GCP_SA_KEY_JSON_B64.</p>"
      exit 1
  fi
  export GOOGLE_APPLICATION_CREDENTIALS="$CRED_FILE"
  echo "Using credentials from $CRED_FILE"
elif [ -d "$HOME/.config/gcloud" ]; then
    echo "Using Application Default Credentials (ADC) found in $HOME/.config/gcloud."
else
    echo "lemc.html.append;<p style='color:red;'>Error: No GCP credentials found. Set GCP_SA_KEY_JSON_B64 or configure ADC.</p>"
    exit 1
fi

# --- Prepare Terraform Variables ---
# Set TF_VAR_* variables from LEMC context and form/env inputs
# Crucial for Terraform to find the correct state file and plan the destroy

# LEMC Context
export TF_VAR_lemc_uuid="${LEMC_UUID:-unknown_uuid}"
export TF_VAR_lemc_username="${LEMC_USERNAME:-unknown_user}"
export TF_VAR_LEMC_SCOPE="${LEMC_SCOPE:-unknown_scope}" # Changed default

# Cookbook Identifier (from public env)
# Use direct variable name COOKBOOK_IDENTIFIER
if [ -n "$COOKBOOK_IDENTIFIER" ]; then
  export TF_VAR_cookbook_identifier="$COOKBOOK_IDENTIFIER"
  echo "Using Cookbook Identifier: $TF_VAR_cookbook_identifier"
else
  export TF_VAR_cookbook_identifier="default-tf-cookbook" # Fallback must match apply if used
  echo "Warning: COOKBOOK_IDENTIFIER not set, using default: $TF_VAR_cookbook_identifier"
fi

# GCP Project ID (Priority: Form/Env > Extracted from SA Key > Error)
# Use direct variable name gcp_project_id
PROJECT_ID="${gcp_project_id:-}"

if [ -z "$PROJECT_ID" ] && [ -f "$CRED_FILE" ]; then
    EXTRACTED_ID=$(grep '"project_id"' "$CRED_FILE" | sed -n 's/.*"project_id": "\([^"]*\)".*/\1/p')
    if [ -n "$EXTRACTED_ID" ]; then
        echo "DEBUG: Extracted Project ID from credentials: $EXTRACTED_ID"
        PROJECT_ID="$EXTRACTED_ID"
    fi
fi

if [ -n "$PROJECT_ID" ]; then
    export TF_VAR_gcp_project_id="$PROJECT_ID"
    echo "Using GCP Project ID: $TF_VAR_gcp_project_id"
else
    echo "lemc.html.append;<p style='color:red;'>Error: GCP Project ID not found. Set 'gcp_project_id' environment variable or provide valid credentials.</p>"
    exit 1
fi

# Optional Vars (Region, Zone, Prefix) - Needed for TF init/plan stages even during destroy
# Use direct variable names gcp_region, gcp_zone, instance_name_prefix
if [ -n "$gcp_region" ]; then export TF_VAR_gcp_region="$gcp_region"; echo "Using GCP Region from Env: $TF_VAR_gcp_region"; else echo "Using default GCP Region."; fi
if [ -n "$gcp_zone" ]; then export TF_VAR_gcp_zone="$gcp_zone"; echo "Using GCP Zone from Env: $TF_VAR_gcp_zone"; else echo "Using default GCP Zone."; fi
if [ -n "$instance_name_prefix" ]; then export TF_VAR_instance_name_prefix="$instance_name_prefix"; echo "Using Prefix from Env: $TF_VAR_instance_name_prefix"; else echo "Using default Instance Name Prefix."; fi

# Domain and Zone Name (Crucial! Read from env vars passed by test script)
# Use direct variable names domain_name, dns_zone_name
if [ -n "$domain_name" ]; then
  export TF_VAR_domain_name="$domain_name"
  echo "Using Domain Name from Env: $TF_VAR_domain_name"
else
  echo "lemc.html.append;<p style='color:red;'>Error: Required variable 'domain_name' not provided.</p>"
  exit 1 # Domain name is likely essential for destroy
fi
if [ -n "$dns_zone_name" ]; then
  export TF_VAR_dns_zone_name="$dns_zone_name"
  echo "Using DNS Zone Name from Env: $TF_VAR_dns_zone_name"
else
  echo "lemc.html.append;<p style='color:red;'>Error: Required variable 'dns_zone_name' not provided.</p>"
  exit 1 # DNS Zone name is likely essential for destroy
fi

# --- Terraform Execution ---
echo "lemc.html.trunc;<h2>Terraform $ACTION_NAME</h2>"
echo "lemc.html.append;<p><b>Context:</b> UUID=$LEMC_UUID, User=$LEMC_USERNAME, Page=$LEMC_PAGE_ID, Recipe=$LEMC_RECIPE_NAME</p>"
echo "lemc.html.append;<p><b>Config:</b> Project=$TF_VAR_gcp_project_id, Region=$TF_VAR_gcp_region, Zone=$TF_VAR_gcp_zone, Prefix=$TF_VAR_instance_name_prefix</p>"

echo "lemc.html.append;<h3>Running Terraform Init...</h3><pre>"
terraform init -no-color || {
  echo "</pre><p style='color:red;'>Terraform Init Failed!</p>"
  exit 1
}
echo "</pre><p>Terraform Init Successful.</p>"

echo "lemc.html.append;<h3>Running Terraform $TF_COMMAND...</h3><pre>"
# Capture output/errors
TF_OUTPUT=$(terraform $TF_COMMAND -no-color 2>&1)
TF_EXIT_CODE=$?
echo "$TF_OUTPUT" # Print the captured output after command finishes
echo "</pre>"

# --- Final Status Reporting ---
if [ $TF_EXIT_CODE -eq 0 ]; then
  echo "lemc.html.append;<hr><p style='color:orange;font-weight:bold;'>Terraform Destroy Completed Successfully!</p>"

  # Clean up working directories inside the container
  echo "Cleaning up container directories /lemc/private and /lemc/public..."
  find /lemc/private -mindepth 1 -delete
  if [ $? -ne 0 ]; then
    echo "Warning: Failed to clean up contents of /lemc/private" >&2
    # Don't exit, just warn
  fi
  find /lemc/public -mindepth 1 -delete
  if [ $? -ne 0 ]; then
    echo "Warning: Failed to clean up contents of /lemc/public" >&2
    # Don't exit, just warn
  fi
  echo "Container directory cleanup complete."

else
  echo "lemc.html.append;<hr><p style='color:red;font-weight:bold;'>Terraform $ACTION_NAME Failed! Exit Code: $TF_EXIT_CODE</p>"
  echo "lemc.html.append;<h5>Last output:</h5><pre>$(echo "$TF_OUTPUT" | tail -n 20)</pre>"
fi

# Clean up credential file if created
if [ -f "$CRED_FILE" ]; then
    rm -f "$CRED_FILE"
fi

echo "--- Entrypoint End (Destroy) ---"
exit $TF_EXIT_CODE 