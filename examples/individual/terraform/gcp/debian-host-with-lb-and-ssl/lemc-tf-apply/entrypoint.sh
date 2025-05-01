#!/bin/sh

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration & Setup ---
TF_WORKING_DIR="/lemc/private"
PUBLIC_DIR="/lemc/public"
CRED_FILE="/tmp/gcp-credentials.json"
EXISTING_KEY_FILE="/tmp/id_rsa_existing.pem"
IMAGE_TF_CONFIG_DIR="/app/tf_config" # Location of files copied in Dockerfile

mkdir -p "$PUBLIC_DIR"
mkdir -p "$TF_WORKING_DIR"
cd "$TF_WORKING_DIR" || exit 1

TF_COMMAND="apply -auto-approve"
ACTION_NAME="Apply"

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
  echo "lemc.html.append;<p style='color:red;font-weight:bold;'>Error: No Terraform files (.tf) found in working directory $TF_WORKING_DIR after attempting copy from image.</p>"
  exit 1
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
    # Terraform should pick these up automatically
else
    echo "lemc.html.append;<p style='color:red;'>Error: No GCP credentials found. Set GCP_SA_KEY_JSON_B64 or configure ADC.</p>"
    exit 1
fi

# --- Existing SSH Key Handling (Optional) ---
# Use direct variable name EXISTING_SSH_KEY_PEM_B64
if [ -n "$EXISTING_SSH_KEY_PEM_B64" ]; then
    echo "Found existing private key in EXISTING_SSH_KEY_PEM_B64."
    echo "Decoding to $EXISTING_KEY_FILE..."
    echo "$EXISTING_SSH_KEY_PEM_B64" | base64 -d > "$EXISTING_KEY_FILE"
    if [ $? -ne 0 ]; then
        echo "lemc.html.append;<p style='color:red;'>Error: Failed to decode EXISTING_SSH_KEY_PEM_B64.</p>"
        exit 1
    fi
    # Validate it looks like a PEM key (basic check)
    if ! grep -q "BEGIN RSA PRIVATE KEY" "$EXISTING_KEY_FILE"; then
        echo "lemc.html.append;<p style='color:red;'>Error: Decoded EXISTING_SSH_KEY_PEM_B64 does not look like a PEM private key.</p>"
        exit 1
    fi
    chmod 600 "$EXISTING_KEY_FILE"
    echo "Exporting TF_VAR_existing_ssh_key_pem with path $EXISTING_KEY_FILE"
    export TF_VAR_existing_ssh_key_pem="$EXISTING_KEY_FILE"
    # Optionally export the public key if needed by TF config (requires ssh-keygen)
    # ssh-keygen -y -f "$EXISTING_KEY_FILE" > "${EXISTING_KEY_FILE}.pub"
    # export TF_VAR_existing_ssh_public_key="$(cat "${EXISTING_KEY_FILE}.pub")"
else
    echo "No existing SSH key provided via EXISTING_SSH_KEY_PEM_B64. Terraform will generate one."
fi


# --- Prepare Terraform Variables ---
# Set TF_VAR_* variables from LEMC context and form/env inputs

# LEMC Context (Always available)
export TF_VAR_lemc_uuid="${LEMC_UUID:-unknown_uuid}"
export TF_VAR_lemc_username="${LEMC_USERNAME:-unknown_user}"
export TF_VAR_lemc_scope="${LEMC_SCOPE:-unknown_scope}" # Changed default

# Cookbook Identifier (from public env)
# Use direct variable name COOKBOOK_IDENTIFIER
if [ -n "$COOKBOOK_IDENTIFIER" ]; then
  export TF_VAR_cookbook_identifier="$COOKBOOK_IDENTIFIER"
  echo "Using Cookbook Identifier: $TF_VAR_cookbook_identifier"
else
  export TF_VAR_cookbook_identifier="default-tf-cookbook" # Fallback
  echo "Warning: COOKBOOK_IDENTIFIER not set, using default: $TF_VAR_cookbook_identifier"
fi

# GCP Project ID (Priority: Form/Env > Extracted from SA Key > Error)
# Use direct variable name gcp_project_id
PROJECT_ID="${gcp_project_id:-}" # Prioritize direct env var

if [ -z "$PROJECT_ID" ] && [ -f "$CRED_FILE" ]; then
    # Attempt to extract from SA JSON if not provided directly
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


# Optional Vars (Region, Zone, Prefix)
# Use direct variable names gcp_region, gcp_zone, instance_name_prefix
if [ -n "$gcp_region" ]; then export TF_VAR_gcp_region="$gcp_region"; echo "Using GCP Region from Env: $TF_VAR_gcp_region"; else echo "Using default GCP Region."; fi
if [ -n "$gcp_zone" ]; then export TF_VAR_gcp_zone="$gcp_zone"; echo "Using GCP Zone from Env: $TF_VAR_gcp_zone"; else echo "Using default GCP Zone."; fi
if [ -n "$instance_name_prefix" ]; then export TF_VAR_instance_name_prefix="$instance_name_prefix"; echo "Using Prefix from Env: $TF_VAR_instance_name_prefix"; else echo "Using default Instance Name Prefix."; fi

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

# --- Terraform Plan ---
echo "lemc.html.append;<h3>Running Terraform plan...</h3><pre>"
terraform plan -no-color || {
    echo "</pre><p style='color:red;'>Terraform Plan Failed!</p>"
    # Optionally exit here if a plan failure should stop the process
    # exit 1
}
echo "</pre><p>Terraform Plan Completed.</p>"


# --- Terraform Apply ---
echo "lemc.html.append;<h3>Running Terraform $TF_COMMAND...</h3><pre>"
# Execute apply directly, streaming output
terraform $TF_COMMAND -no-color
TF_EXIT_CODE=$?
echo "</pre>" # Close the <pre> tag after the command finishes

# --- Post-Apply Steps ---
if [ $TF_EXIT_CODE -eq 0 ]; then
  echo "lemc.html.append;<hr><p style='color:green;font-weight:bold;'>Terraform Apply Completed Successfully!</p>"
  
  # --- Retrieve outputs ---
  echo "lemc.html.append;<h4>Retrieving Outputs...</h4>"
  PUBLIC_IP=$(terraform output -raw public_ip 2>/dev/null || echo "N/A")
  INSTANCE_NAME=$(terraform output -raw instance_name 2>/dev/null || echo "N/A")
  PUBLIC_KEY_OUTPUT=$(terraform output -raw public_ssh_key 2>/dev/null || echo "N/A")
  PRIVATE_KEY_OUTPUT=$(terraform output -raw private_ssh_key 2>/dev/null || echo "N/A") # Sensitive!
  
  # --- Display basic outputs ---
  echo "lemc.html.append;<h5>Outputs:</h5><ul>"
  echo "lemc.html.append;<li>Instance Name: <code>$INSTANCE_NAME</code></li>"
  echo "lemc.html.append;<li>Public IP: <code>$PUBLIC_IP</code></li>"
  echo "lemc.html.append;<li>Public Key (first 40 chars): <code>$(echo "$PUBLIC_KEY_OUTPUT" | cut -c 1-40)...</code></li>"
  echo "lemc.html.append;</ul>"
  
  # --- Save private key and provide download link ---
  if [ "$PRIVATE_KEY_OUTPUT" != "N/A" ] && [ -n "$PRIVATE_KEY_OUTPUT" ]; then
      KEY_FILENAME="id_rsa_${LEMC_USERNAME}_${LEMC_UUID}.pem"
      KEY_PATH="$PUBLIC_DIR/$KEY_FILENAME"
      echo "Saving private key to $KEY_PATH..."
      echo "$PRIVATE_KEY_OUTPUT" > "$KEY_PATH"
      chmod 600 "$KEY_PATH"
      
      # Check if file exists and is readable before creating link
      if [ -r "$KEY_PATH" ]; then
          DOWNLOAD_URL="${LEMC_HTTP_DOWNLOAD_BASE_URL}${KEY_FILENAME}"
          echo "lemc.html.append;<p><b>SSH Private Key generated:</b> <a href='$DOWNLOAD_URL' target='_blank' download='$KEY_FILENAME' style='font-weight:bold; color:orange;'>Download $KEY_FILENAME</a></p>"
          echo "lemc.html.append;<p>Use this key to connect: <code>ssh -i /path/to/$KEY_FILENAME ${LEMC_USERNAME}@${PUBLIC_IP}</code></p>"
      else
           echo "lemc.html.append;<p style='color:red;'>Error: Could not save or read private key file at $KEY_PATH.</p>"
      fi
  elif [ -n "$EXISTING_SSH_KEY_PEM_B64" ]; then
      echo "lemc.html.append;<p>Using existing SSH key provided. No new key generated.</p>"
  else
      echo "lemc.html.append;<p style='color:orange;'>Warning: Could not retrieve private key from Terraform output.</p>"
  fi
else
  # --- Final Status Reporting for failure ---
  echo "lemc.html.append;<hr><p style='color:red;font-weight:bold;'>Terraform $ACTION_NAME Failed! Exit Code: $TF_EXIT_CODE</p>"
  # We no longer have TF_OUTPUT, so we can't show the last lines easily here.
  # You might need to capture logs differently if you need this specific feature back.
fi

# Clean up credential file if created
if [ -f "$CRED_FILE" ]; then
    rm -f "$CRED_FILE"
fi
if [ -f "$EXISTING_KEY_FILE" ]; then
    rm -f "$EXISTING_KEY_FILE"
fi

exit $TF_EXIT_CODE