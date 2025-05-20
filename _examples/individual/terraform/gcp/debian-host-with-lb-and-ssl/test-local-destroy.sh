#!/bin/sh
set -a # Automatically export all variables
set -e # Exit on error

# --- Local Test Harness for lemc-tf-destroy ---
# This script mimics the LEMC environment for local testing.

# --- Source Environment Variables from dotenv file ---
if [ -f dotenv ]; then
  echo "INFO: Sourcing environment variables from dotenv file..."
  # Use '.' to source in the current shell context
  . ./dotenv
else
  echo "WARNING: dotenv file not found. Using default script values or environment variables."
fi

# --- ========================================= ---
# --- Section 1: Environment Variable Setup     ---
# --- ========================================= ---

# --- Mock System LEMC Context Variables (Defaults if not set in dotenv) ---
# IMPORTANT: For destroy, LEMC_UUID should ideally match the apply run you want to destroy!
: "${LEMC_UUID:=mock-lemc-uuid-$(date +%s)}" # Default UUID includes timestamp
: "${LEMC_USERNAME:=mock-user}"
: "${LEMC_SCOPE:=individual}"

# --- User-Defined Root Domain/Zone (Defaults if not set in dotenv) ---
: "${ROOT_DOMAIN:=example.com}"
: "${ROOT_ZONE:=my-gcp-dns-zone-name}"

# --- Required GCP Project ID (Check & Default) ---
# GCP Project ID (Required - Check after sourcing dotenv)
: "${GCP_PROJECT_ID:?Error: GCP_PROJECT_ID must be set (in dotenv or environment)}"

# --- Construct Dynamic Names from LEMC &  Vars ---
LEMC_UUID_SHORT=$(echo "$LEMC_UUID" | cut -c 1-8)
DYNAMIC_DOMAIN_NAME="${LEMC_UUID_SHORT}-${LEMC_USERNAME}-${LEMC_SCOPE}.${ROOT_DOMAIN}"
DYNAMIC_DNS_ZONE_NAME="${ROOT_ZONE}"

# --- Terraform Variables (Exported for Terraform) ---

# Inject LEMC context directly as VARs
export VAR_lemc_uuid="$LEMC_UUID"
export VAR_lemc_username="$LEMC_USERNAME"
export VAR_LEMC_SCOPE="$LEMC_SCOPE"

# Export the dynamically constructed domain/zone names
export VAR_domain_name="${DYNAMIC_DOMAIN_NAME}"
export VAR_dns_zone_name="${DYNAMIC_DNS_ZONE_NAME}"

# Export optional  prefixed variables directly for Terraform
[ -n "$COOKBOOK_IDENTIFIER" ] && export VAR_COOKBOOK_IDENTIFIER="${COOKBOOK_IDENTIFIER}"
[ -n "$INSTANCE_NAME_PREFIX" ] && export VAR_INSTANCE_NAME_PREFIX="${INSTANCE_NAME_PREFIX}"
[ -n "$GCP_REGION" ] && export VAR_GCP_REGION="${GCP_REGION}"
[ -n "$GCP_ZONE" ] && export VAR_GCP_ZONE="${GCP_ZONE}"

# --- Optional Cookbook/Form Variables (Defaults if not set in dotenv) --- (ALL CAPS,  prefix)
# Use direct names, no prefixes, as they will be passed to the container
: "${COOKBOOK_IDENTIFIER:=local-tf-demo}"
: "${INSTANCE_NAME_PREFIX:=test-prefix}"
: "${GCP_REGION:=us-central1}"
: "${GCP_ZONE:=us-central1-a}"

# --- Configuration (Defaults if not set in dotenv) --- (ALL CAPS)
: "${IMAGE_NAME:=${IMAGE_NAME_DESTROY:-docker.io/jfolkins/lemc-tf-destroy:latest}}" # Use IMAGE_NAME_DESTROY from dotenv if set

# Mock Directories (relative to this script's location)
: "${PRIVATE_DIR:=./private}"
: "${PUBLIC_DIR:=./public}" # Still needed for potential key cleanup

# --- ========================================= ---
# --- Section 2: Initialization & Checks      ---
# --- ========================================= ---

# --- Argument Parsing --- (For GCP Key File)
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

# --- Instructions --- (Displayed early)
echo "--- Terraform Destroy Local Test --- "
echo "1. Ensure Docker image '$IMAGE_NAME' is built locally."
echo "2. Ensure GCP Project ID is set via GCP_PROJECT_ID environment variable."
echo "3. Provide GCP credentials via one of these methods:"
echo "   a) Use the '-k /path/to/key.json' or '--key-file /path/to/key.json' option."
echo "   b) Ensure you are authenticated with GCP locally ('gcloud auth application-default login')."
echo "4. Ensure the '${PRIVATE_DIR}' directory exists and contains the 'terraform.tfstate' file"
echo "   from a previous apply run. Ensure '${PUBLIC_DIR}' also exists."
echo "   Note: Terraform configuration files (.tf) are now included in the Docker image."
echo "------------------------------------"

# --- Check for Required Tools (jq, base64) ---
if ! command -v jq > /dev/null 2>&1; then
    echo "ERROR: \`jq\` command not found. Please install jq." >&2
    exit 1
fi
if ! command -v base64 > /dev/null 2>&1; then
    echo "ERROR: \`base64\` command not found. Please install it." >&2
    exit 1
fi

# --- Check Mock Dirs & State File ---
if [ ! -d "${PRIVATE_DIR}" ]; then
    echo "ERROR: Mock private directory '${PRIVATE_DIR}' not found. Run apply test first?" >&2
    exit 1
fi
if [ ! -f "${PRIVATE_DIR}/terraform.tfstate" ]; then
    echo "WARNING: No terraform.tfstate file found in ${PRIVATE_DIR}. Destroy will likely fail or do nothing."
fi
mkdir -p "${PUBLIC_DIR}"

# --- Display Derived Variables --- (Moved here after calculation)
echo "==> Using derived variables for destroy:"
echo "    VAR_domain_name: ${VAR_domain_name}"
echo "    VAR_dns_zone_name: ${VAR_dns_zone_name}"
echo "-----------------------------------------"

# --- GCP Credentials Handling --- (Prepare for Docker injection)
# Priority: 1. Key File Path (-k) -> 2. gcloud ADC
EXTRACTED_GCP_PROJECT_ID="" # We already checked GCP_PROJECT_ID
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
  echo "ERROR: No GCP credentials provided. Please use -k <file> or run 'gcloud auth application-default login'." >&2
  exit 1
fi

# --- ========================================= ---
# --- Section 3: Docker Execution             ---
# --- ========================================= ---

# --- Docker Command Assembly --- (Assemble just before running)
docker_cmd=(docker run --rm --name "lemc-tf-destroy-local-test" --hostname "lemc-tf-destroy-local-test")

# Volumes
docker_cmd+=("-v" "$(pwd)/${PRIVATE_DIR}:/lemc/private") # Mount private working dir (needs state file)
docker_cmd+=("-v" "$(pwd)/${PUBLIC_DIR}:/lemc/public")   # Mount public output dir (usually empty for destroy)

# Inject Credentials if using SA Key (as Base64)
if [ -n "$GCP_SA_KEY_JSON_B64" ]; then
  docker_cmd+=("-e" "GCP_SA_KEY_JSON_B64=${GCP_SA_KEY_JSON_B64}")
fi
# Note: If using ADC, the entrypoint script should handle mounting gcloud config

# Inject System LEMC Context Variables (Already exported, but pass explicitly for clarity)
docker_cmd+=("-e" "LEMC_SCOPE=${LEMC_SCOPE}")
# docker_cmd+=("-e" "LEMC_USER_ID=${LEMC_USER_ID}") # Not typically needed by destroy entrypoint
docker_cmd+=("-e" "LEMC_USERNAME=${LEMC_USERNAME}")
docker_cmd+=("-e" "LEMC_UUID=${LEMC_UUID}")
# docker_cmd+=("-e" "LEMC_RECIPE_NAME=${LEMC_RECIPE_NAME}") # Not relevant for destroy
# docker_cmd+=("-e" "LEMC_PAGE_ID=${LEMC_PAGE_ID}")         # Not relevant for destroy
# No download base URL needed for destroy typically

# Inject Cookbook/Form Variables (passing  prefixed values using original lowercase names, needed for TF vars)
docker_cmd+=("-e" "COOKBOOK_IDENTIFIER=${COOKBOOK_IDENTIFIER}")
docker_cmd+=("-e" "gcp_project_id=${GCP_PROJECT_ID}") # Pass GCP_PROJECT_ID as gcp_project_id

# Add the required domain and DNS zone variables (without TF_VAR_ prefix based on user feedback)
docker_cmd+=("-e" "domain_name=${VAR_domain_name}")
docker_cmd+=("-e" "dns_zone_name=${VAR_dns_zone_name}")

# Optional variables (only add if they have values - needed for Terraform init/plan)
# Pass these as lowercase, consistent with gcp_project_id
[ -n "$GCP_REGION" ] && docker_cmd+=("-e" "gcp_region=${GCP_REGION}")
[ -n "$GCP_ZONE" ] && docker_cmd+=("-e" "gcp_zone=${GCP_ZONE}")
[ -n "$INSTANCE_NAME_PREFIX" ] && docker_cmd+=("-e" "instance_name_prefix=${INSTANCE_NAME_PREFIX}")

# Image Name
docker_cmd+=("$IMAGE_NAME")

# --- Execute --- (The final action)
echo "Running Terraform Destroy container..."
echo "Executing: ${docker_cmd[*]}"
echo "--- Container Output Start ---"

# Execute the command directly without eval
"${docker_cmd[@]}"
EXIT_CODE=$?

echo "--- Container Output End --- (Exit Code: $EXIT_CODE)"

if [ $EXIT_CODE -eq 0 ]; then
  echo "Destroy test finished successfully."
  echo "Check '${PRIVATE_DIR}' - state file might be removed or updated by Terraform."
else
  echo "Destroy test finished with errors (Exit Code: $EXIT_CODE)."
  # Exit the test script with the container's non-zero exit code
  exit $EXIT_CODE
fi

# Unset exported variables if needed (though script exits)
set +a 