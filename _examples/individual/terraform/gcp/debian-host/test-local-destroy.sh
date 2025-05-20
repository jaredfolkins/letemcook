#!/bin/sh
set -a # Automatically export all variables
set -e # Exit on error

# --- Local Test Harness for lemc-tf-destroy ---
# This script mimics the LEMC environment for local testing.

# --- Configuration ---
# Replace with the actual image name you built
IMAGE_NAME="docker.io/jfolkins/lemc-tf-destroy:latest"

# Mock Directories (relative to this script's location)
# IMPORTANT: This should point to the same directory used by the apply test harness
# as it needs the terraform.tfstate file created by the apply run.
MOCK_PRIVATE_DIR="./mock-lemc-private"
MOCK_PUBLIC_DIR="./mock-lemc-public" # Still needed for potential key cleanup

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

# --- Instructions ---
echo "--- Terraform Destroy Local Test --- "
echo "1. Ensure Docker image '$IMAGE_NAME' is built locally."
echo "2. Provide GCP credentials via one of these methods:"
echo "   a) Use the '-k /path/to/key.json' or '--key-file /path/to/key.json' option."
echo "   b) Set the LEMC_PRIVATE_GCP_SA_KEY_JSON environment variable."
echo "   c) Ensure you are authenticated with GCP locally ('gcloud auth application-default login')."
echo "3. Ensure the '${MOCK_PRIVATE_DIR}' directory exists and contains the 'terraform.tfstate' file"
echo "   from a previous apply run. Ensure '${MOCK_PUBLIC_DIR}' also exists."
echo "   Note: Terraform configuration files (.tf) are now included in the Docker image."
echo "------------------------------------"

# --- Check Mock Dirs ---
if [ ! -d "${MOCK_PRIVATE_DIR}" ]; then
    echo "ERROR: Mock private directory '${MOCK_PRIVATE_DIR}' not found. Run apply test first?" >&2
    exit 1
fi
# Remove the check for .tf files
# if ! ls "${MOCK_PRIVATE_DIR}"/*.tf > /dev/null 2>&1; then
#  echo "ERROR: No *.tf files found in ${MOCK_PRIVATE_DIR}. Copy them there." >&2
#  exit 1
# fi
if [ ! -f "${MOCK_PRIVATE_DIR}/terraform.tfstate" ]; then
    echo "WARNING: No terraform.tfstate file found in ${MOCK_PRIVATE_DIR}. Destroy will likely fail or do nothing."
fi
# Also ensure public dir exists for potential key cleanup
mkdir -p "${MOCK_PUBLIC_DIR}"

# --- Check for jq and base64 ---\nif ! command -v jq > /dev/null 2>&1; then\n    echo \"ERROR: \`jq\` command not found. Please install jq.\" >&2\n    exit 1\nfi\nif ! command -v base64 > /dev/null 2>&1; then\n    echo \"ERROR: \`base64\` command not found. Please install it.\" >&2\n    exit 1\nfi\n

# --- Mock LEMC Environment Variables ---
# These should generally match the context of the apply run you want to destroy
# Basic Context
LEMC_UUID="mockuuid-1234-5678-9abc-def012345678" # Should match the apply run UUID for correct naming/state lookup
LEMC_USERNAME="local-mock-user" # Should match the apply run user for consistency
LEMC_RECIPE_NAME="local-destroy-recipe"
LEMC_PAGE_ID="1"
LEMC_SCOPE="individual"

# --- GCP Credentials Handling ---
# Priority: 1. Key File Path (-k) -> 2. gcloud ADC
EXTRACTED_GCP_PROJECT_ID=""
# Renamed to match expected var name
GCP_SA_KEY_JSON_B64="" # Base64 encoded key

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
        echo "WARNING: Could not automatically extract project_id from key file $GCP_KEY_FILE_PATH (jq exit code: $JQ_EXIT_CODE). Setting EXTRACTED_GCP_PROJECT_ID to empty." >&2
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

# --- Cookbook / Form Variables (Must match the context expected by the state file) ---
# Mimic variables from LEMC forms or cookbook environment
# Use direct names, no prefixes
COOKBOOK_IDENTIFIER="local-tf-demo" # Corresponds to TF_VAR_cookbook_identifier
gcp_project_id=""                 # Corresponds to TF_VAR_gcp_project_id
instance_name_prefix="test-prefix"  # Optional, Corresponds to TF_VAR_instance_name_prefix
gcp_region="us-central1"            # Optional, Corresponds to TF_VAR_gcp_region
gcp_zone="us-central1-a"              # Optional, Corresponds to TF_VAR_gcp_zone

# --- Determine GCP Project ID ---
# Priority: Command line > Environment > ADC extracted > Default
if [ -z "$gcp_project_id" ]; then
  if [ -n "${TF_VAR_gcp_project_id-}" ]; then # Check common TF env var too
      gcp_project_id="$TF_VAR_gcp_project_id"
      echo "Using GCP Project ID from TF_VAR_gcp_project_id env var: $gcp_project_id"
  elif gcloud config get-value project > /dev/null 2>&1; then
      gcp_project_id=$(gcloud config get-value project)
      echo "Using GCP Project ID from gcloud config: $gcp_project_id"
  else
      gcp_project_id="your-gcp-project-id-needs-setting" # Fallback default
      echo "Warning: GCP Project ID not set via -p, TF_VAR_gcp_project_id, or gcloud config. Using placeholder: $gcp_project_id"
  fi
fi

# --- Docker Command Assembly ---
docker_cmd=(docker run --rm --name "lemc-tf-destroy-local-test" --hostname "lemc-tf-destroy-local-test")

# Volumes
docker_cmd+=("-v" "$(pwd)/mock-lemc-private:/lemc/private") # Mount private working dir (needs state file)
docker_cmd+=("-v" "$(pwd)/mock-lemc-public:/lemc/public")   # Mount public output dir (usually empty for destroy)

# Inject Credentials if using SA Key (as Base64)
if [ -n "$GCP_SA_KEY_JSON_B64" ]; then
  docker_cmd+=("-e" "GCP_SA_KEY_JSON_B64=${GCP_SA_KEY_JSON_B64}")
fi
# Note: If using ADC, the entrypoint script should handle mounting gcloud config

# Inject LEMC Context Variables
docker_cmd+=("-e" "LEMC_SCOPE=${LEMC_SCOPE}")
docker_cmd+=("-e" "LEMC_USER_ID=${LEMC_USER_ID}")
docker_cmd+=("-e" "LEMC_USERNAME=${LEMC_USERNAME}")
docker_cmd+=("-e" "LEMC_UUID=${LEMC_UUID}")
docker_cmd+=("-e" "LEMC_RECIPE_NAME=${LEMC_RECIPE_NAME}") # Using mocked value
docker_cmd+=("-e" "LEMC_PAGE_ID=${LEMC_PAGE_ID}")         # Using mocked value
# No download base URL needed for destroy typically

# Inject Cookbook/Form Variables (without prefixes)
docker_cmd+=("-e" "COOKBOOK_IDENTIFIER=${COOKBOOK_IDENTIFIER}")
docker_cmd+=("-e" "gcp_project_id=${gcp_project_id}")

# Optional variables (only add if they have values - needed for Terraform init/plan)
[ -n "$gcp_region" ] && docker_cmd+=("-e" "gcp_region=${gcp_region}")
[ -n "$gcp_zone" ] && docker_cmd+=("-e" "gcp_zone=${gcp_zone}")
[ -n "$instance_name_prefix" ] && docker_cmd+=("-e" "instance_name_prefix=${instance_name_prefix}")

# Image Name
docker_cmd+=("$IMAGE_NAME")

# --- Execute ---
echo "Running Terraform Destroy container..."
echo "Executing: ${docker_cmd[*]}"
echo "--- Container Output Start ---"

# Execute the command directly without eval
"${docker_cmd[@]}"
EXIT_CODE=$?

echo "--- Container Output End --- (Exit Code: $EXIT_CODE)"

if [ $EXIT_CODE -eq 0 ]; then
  echo "Destroy test finished successfully."
  echo "Check '${MOCK_PRIVATE_DIR}' - state file might be removed or updated by Terraform."
  echo "Check '${MOCK_PUBLIC_DIR}' - SSH keys should be removed."
else
  echo "Destroy test finished with errors (Exit Code: $EXIT_CODE)."
  # Exit the test script with the container's non-zero exit code
  exit $EXIT_CODE
fi

# Unset exported variables if needed
set +a 