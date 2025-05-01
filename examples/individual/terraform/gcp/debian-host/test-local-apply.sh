#!/bin/sh
set -a # Automatically export all variables
set -e # Exit on error

# --- Local Test Harness for lemc-tf-apply ---
# This script mimics the LEMC environment for local testing.

# --- Configuration ---
# Replace with the actual image name you built
IMAGE_NAME="docker.io/jfolkins/lemc-tf-apply:latest"

# Mock Directories (relative to this script's location)
MOCK_PRIVATE_DIR="./mock-lemc-private"
MOCK_PUBLIC_DIR="./mock-lemc-public"

# --- Argument Parsing ---
GCP_KEY_FILE_PATH=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -k|--key-file)
      if [ -n "$2" ]; then
        GCP_KEY_FILE_PATH="$2"
        shift 2
      else
        echo "Error: --key-file requires a path argument." >&2
        exit 1
      fi
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

# --- Check for jq and base64 ---\nif ! command -v jq > /dev/null 2>&1; then\n    echo \"ERROR: \`jq\` command not found. Please install jq.\" >&2\n    exit 1\nfi\nif ! command -v base64 > /dev/null 2>&1; then\n    echo \"ERROR: \`base64\` command not found. Please install it.\" >&2\n    exit 1\nfi\n\n# --- Instructions ---
echo "--- Terraform Apply Local Test --- "
echo "1. Ensure Docker image '$IMAGE_NAME' is built locally."
echo "2. Provide GCP credentials via one of these methods:"
echo "   a) Use the '-k /path/to/key.json' or '--key-file /path/to/key.json' option."
echo "   b) Set the LEMC_PRIVATE_GCP_SA_KEY_JSON environment variable."
echo "   c) Ensure you are authenticated with GCP locally ('gcloud auth application-default login')."
echo "3. Ensure the '${MOCK_PRIVATE_DIR}' and '${MOCK_PUBLIC_DIR}' directories exist (they will be created if missing)."
echo "   Note: Terraform configuration files (.tf) are now included in the Docker image."
echo "---------------------------------"

# --- Create Mock Dirs ---
mkdir -p "${MOCK_PRIVATE_DIR}"
mkdir -p "${MOCK_PUBLIC_DIR}"

# --- Mock LEMC Environment Variables ---
# Basic Context
LEMC_UUID="mockuuid-1234-5678-9abc-def012345678"
LEMC_USERNAME="local-mock-user"
LEMC_RECIPE_NAME="local-mock-apply-recipe"
LEMC_PAGE_ID="1"
LEMC_SCOPE="individual"
LEMC_HTTP_DOWNLOAD_BASE_URL="/mock-download-path" # Needed for download link generation

# --- GCP Credentials Handling ---
# Priority: 1. Key File Path (-k) -> 2. gcloud ADC
EXTRACTED_GCP_PROJECT_ID=""
# Renamed to match the expected variable after prefix removal
GCP_SA_KEY_JSON_B64="" # Encoded key

if [ -n "$GCP_KEY_FILE_PATH" ]; then
    echo "INFO: Using key file specified by -k/--key-file: $GCP_KEY_FILE_PATH"
    if [ ! -f "$GCP_KEY_FILE_PATH" ]; then
        echo "ERROR: Specified GCP key file does not exist: $GCP_KEY_FILE_PATH" >&2
        exit 1
    fi
    if [ ! -r "$GCP_KEY_FILE_PATH" ]; then
        echo "ERROR: Specified GCP key file is not readable: $GCP_KEY_FILE_PATH" >&2
        exit 1
    fi
    # Attempt to extract project_id using jq
    EXTRACTED_GCP_PROJECT_ID=$(jq -r '.project_id' "$GCP_KEY_FILE_PATH")
    JQ_EXIT_CODE=$?
    if [ $JQ_EXIT_CODE -ne 0 ] || [ -z "$EXTRACTED_GCP_PROJECT_ID" ] || [ "$EXTRACTED_GCP_PROJECT_ID" = "null" ]; then
        echo "WARNING: Could not automatically extract project_id from key file $GCP_KEY_FILE_PATH (jq exit code: $JQ_EXIT_CODE)." >&2
        EXTRACTED_GCP_PROJECT_ID=""
    else
        echo "INFO: Extracted project_id '$EXTRACTED_GCP_PROJECT_ID' from key file."
    fi

    # Read the full key content and base64 encode it
    GCP_SA_KEY_JSON_B64=$(base64 < "$GCP_KEY_FILE_PATH")
    if [ $? -ne 0 ] || [ -z "$GCP_SA_KEY_JSON_B64" ]; then
        echo "ERROR: Failed to Base64 encode key file: $GCP_KEY_FILE_PATH" >&2
        exit 1
    fi
    echo "INFO: Base64 encoded GCP credentials from file."

elif gcloud auth application-default print-access-token --quiet > /dev/null 2>&1; then
  echo "INFO: Using local gcloud Application Default Credentials."
else
  echo "ERROR: No GCP credentials provided. Please use -k <file>, set LEMC_PRIVATE_GCP_SA_KEY_JSON, or run 'gcloud auth application-default login'." >&2
  exit 1
fi

# --- Cookbook / Form Variables (adjust as needed) ---
# Priority: LEMC_FORM > LEMC_PUBLIC > Extracted from Key > Default
# Renamed to match expected variable names after prefix removal
COOKBOOK_IDENTIFIER="local-tf-demo" # Corresponds to TF_VAR_cookbook_identifier
gcp_project_id=""                 # Corresponds to TF_VAR_gcp_project_id
instance_name_prefix="test-prefix"  # Corresponds to TF_VAR_instance_name_prefix
gcp_region="us-central1"            # Corresponds to TF_VAR_gcp_region
gcp_zone="us-central1-a"              # Corresponds to TF_VAR_gcp_zone
# Optional: Existing SSH Key (Base64 encoded PEM)
# EXISTING_SSH_KEY_PEM_B64="" # Corresponds to TF_VAR_existing_ssh_key_pem

# Set Project ID: Use explicit form var > extracted var > default/placeholder
if [ -z "${LEMC_FORM_GCP_PROJECT_ID-}" ] && [ -n "$EXTRACTED_GCP_PROJECT_ID" ]; then
    LEMC_FORM_GCP_PROJECT_ID="$EXTRACTED_GCP_PROJECT_ID"
elif [ -z "${LEMC_FORM_GCP_PROJECT_ID-}" ]; then
    # If not set by env/form and not extracted, use placeholder (will cause error later if not overridden by ADC)
    LEMC_FORM_GCP_PROJECT_ID="your-gcp-project-id-needs-setting"
fi

# --- Build docker run command args ---
# Initialize the command array
docker_cmd=(
    docker run \
    --rm --init \
)

# --- Add environment variables EXPLICITLY ---
# Base64 encoded key (if set)
if [ -n "$GCP_SA_KEY_JSON_B64" ]; then
    docker_cmd+=("-e" "GCP_SA_KEY_JSON_B64=$GCP_SA_KEY_JSON_B64")
fi

# Explicitly add required/known LEMC variables
docker_cmd+=("-e" "LEMC_UUID=${LEMC_UUID}")
docker_cmd+=("-e" "LEMC_USERNAME=${LEMC_USERNAME}")
docker_cmd+=("-e" "LEMC_RECIPE_NAME=${LEMC_RECIPE_NAME}")
docker_cmd+=("-e" "LEMC_PAGE_ID=${LEMC_PAGE_ID}")
docker_cmd+=("-e" "LEMC_SCOPE=${LEMC_SCOPE}")
docker_cmd+=("-e" "LEMC_HTTP_DOWNLOAD_BASE_URL=${LEMC_HTTP_DOWNLOAD_BASE_URL}")
docker_cmd+=("-e" "LEMC_PUBLIC_COOKBOOK_IDENTIFIER=${COOKBOOK_IDENTIFIER}")
docker_cmd+=("-e" "LEMC_FORM_GCP_PROJECT_ID=${LEMC_FORM_GCP_PROJECT_ID}")

# Add optional form variables if they are set in the script environment
[ -n "$gcp_region" ] && docker_cmd+=("-e" "gcp_region=${gcp_region}")
[ -n "$gcp_zone" ] && docker_cmd+=("-e" "gcp_zone=${gcp_zone}")
[ -n "$instance_name_prefix" ] && docker_cmd+=("-e" "instance_name_prefix=${instance_name_prefix}")

# Add Volume Mounts
# Construct absolute paths
ABS_MOCK_PRIVATE_DIR="$(pwd)/${MOCK_PRIVATE_DIR}"
ABS_MOCK_PUBLIC_DIR="$(pwd)/${MOCK_PUBLIC_DIR}"
# Add Volume Mounts as separate arguments
docker_cmd+=("-v" "${ABS_MOCK_PRIVATE_DIR}:/lemc/private")
docker_cmd+=("-v" "${ABS_MOCK_PUBLIC_DIR}:/lemc/public")

# Add Image Name
docker_cmd+=("$IMAGE_NAME")

# Print the command for debugging
echo "Executing: ${docker_cmd[@]}"
echo "--- Container Output Start ---"

# Execute the command directly without eval
"${docker_cmd[@]}"
EXIT_CODE=$?

echo "--- Container Output End --- (Exit Code: $EXIT_CODE)"

if [ $EXIT_CODE -eq 0 ]; then
  echo "Apply test finished successfully."
  echo "Check '${MOCK_PRIVATE_DIR}' for terraform state."
  echo "Check '${MOCK_PUBLIC_DIR}' for the generated private key (if apply succeeded)."
else
  echo "Apply test finished with errors (Exit Code: $EXIT_CODE)."
  # Exit the test script with the container's non-zero exit code
  exit $EXIT_CODE
fi

# Unset exported variables if needed (though script exits)
set +a 
