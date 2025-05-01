#!/bin/sh
set -a # Automatically export all variables
set -e # Exit on error

# --- Local Test Harness for lemc-tf-apply ---
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
: "${LEMC_UUID:=mock-lemc-uuid-$(date +%s)}" # Default UUID includes timestamp if not set
: "${LEMC_USERNAME:=mock-user}"
: "${LEMC_SCOPE:=individual}"
: "${LEMC_RECIPE_NAME:=local-mock-apply-recipe}"
: "${LEMC_PAGE_ID:=1}"
: "${LEMC_HTTP_DOWNLOAD_BASE_URL:=/lemc/locker/uuid/${LEMC_UUID}/page/${LEMC_PAGE_ID:-1}/scope/${LEMC_SCOPE}/filename/}" # Default Mock URL

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
: "${IMAGE_NAME:=${IMAGE_NAME_APPLY:-docker.io/jfolkins/lemc-tf-apply:latest}}" # Use IMAGE_NAME_APPLY from dotenv if set
: "${PRIVATE_DIR:=./private}"
: "${PUBLIC_DIR:=./public}"

# Define path for private key relative to PUBLIC_DIR
PRIVATE_KEY_FILENAME="id_rsa_${LEMC_USERNAME}_${LEMC_UUID}.pem"
PRIVATE_KEY_PATH="${PUBLIC_DIR}/${PRIVATE_KEY_FILENAME}"

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

# --- Check for Required Tools (jq, base64) ---
if ! command -v jq > /dev/null 2>&1; then
    echo "ERROR: \`jq\` command not found. Please install jq." >&2
    exit 1
fi
if ! command -v base64 > /dev/null 2>&1; then
    echo "ERROR: \`base64\` command not found. Please install it." >&2
    exit 1
fi

# --- Instructions --- (Displayed early)
echo "--- Terraform Apply Local Test --- "
echo "1. Ensure Docker image '$IMAGE_NAME' is built locally."
echo "2. Ensure GCP Project ID is set via GCP_PROJECT_ID environment variable."
echo "3. Provide GCP credentials via one of these methods:"
echo "   a) Use the '-k /path/to/key.json' or '--key-file /path/to/key.json' option."
echo "   b) Ensure you are authenticated with GCP locally ('gcloud auth application-default login')."
echo "4. Ensure the '${PRIVATE_DIR}' and '${PUBLIC_DIR}' directories exist (they will be created if missing)."
echo "   Note: Terraform configuration files (.tf) are now included in the Docker image."
echo "---------------------------------"

# --- Create Mock Dirs ---
mkdir -p "${PRIVATE_DIR}"
mkdir -p "${PUBLIC_DIR}"

# --- Display Derived Variables --- (Moved here after calculation)
echo "==> Using derived variables:"
echo "    VAR_domain_name: ${VAR_domain_name}"
echo "    VAR_dns_zone_name: ${VAR_dns_zone_name}"
echo "-----------------------------------------"

# --- GCP Credentials Handling --- (Prepare for Docker injection)
# Priority: 1. Key File Path (-k) -> 2. gcloud ADC
EXTRACTED_GCP_PROJECT_ID="" # Not strictly needed here as we check GCP_PROJECT_ID
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
docker_cmd=(docker run --rm --init --name "lemc-tf-apply-local-test" --hostname "lemc-tf-apply-local-test")

# Add environment variables EXPLICITLY
# Base64 encoded key (if set)
if [ -n "$GCP_SA_KEY_JSON_B64" ]; then
    docker_cmd+=("-e" "GCP_SA_KEY_JSON_B64=$GCP_SA_KEY_JSON_B64")
fi

# Explicitly add required/known System LEMC variables
docker_cmd+=("-e" "LEMC_UUID=${LEMC_UUID}")
docker_cmd+=("-e" "LEMC_USERNAME=${LEMC_USERNAME}")
docker_cmd+=("-e" "LEMC_RECIPE_NAME=${LEMC_RECIPE_NAME}")
docker_cmd+=("-e" "LEMC_PAGE_ID=${LEMC_PAGE_ID}")
docker_cmd+=("-e" "LEMC_SCOPE=${LEMC_SCOPE}")
docker_cmd+=("-e" "LEMC_HTTP_DOWNLOAD_BASE_URL=${LEMC_HTTP_DOWNLOAD_BASE_URL}")

# Add Cookbook/Form variables (passing prefixed values using original lowercase names)
# These need the  prefix for Terraform to automatically recognize them
docker_cmd+=("-e" "COOKBOOK_IDENTIFIER=${COOKBOOK_IDENTIFIER}")
docker_cmd+=("-e" "gcp_project_id=${GCP_PROJECT_ID}") # Pass GCP_PROJECT_ID as gcp_project_id

# Add the required domain and DNS zone variables
# These also need the  prefix
docker_cmd+=("-e" "domain_name=${VAR_domain_name}")
docker_cmd+=("-e" "dns_zone_name=${VAR_dns_zone_name}")

# Add optional variables if they are set in the script environment (passing prefixed values using original lowercase names)
# These also need the  prefix
[ -n "$GCP_REGION" ] && docker_cmd+=("-e" "GCP_REGION=${GCP_REGION}")
[ -n "$GCP_ZONE" ] && docker_cmd+=("-e" "GCP_ZONE=${GCP_ZONE}")
[ -n "$INSTANCE_NAME_PREFIX" ] && docker_cmd+=("-e" "INSTANCE_NAME_PREFIX=${INSTANCE_NAME_PREFIX}")

# Add Volume Mounts
# Construct absolute paths
ABS_PRIVATE_DIR="$(pwd)/${PRIVATE_DIR}"
ABS_PUBLIC_DIR="$(pwd)/${PUBLIC_DIR}"
# Add Volume Mounts as separate arguments
docker_cmd+=("-v" "${ABS_PRIVATE_DIR}:/lemc/private")
docker_cmd+=("-v" "${ABS_PUBLIC_DIR}:/lemc/public")

# Add Image Name
docker_cmd+=("$IMAGE_NAME")

# Print the command for debugging
echo "Executing: ${docker_cmd[*]}"
echo "--- Container Output Start ---"

# Execute the command directly without eval
"${docker_cmd[@]}"
EXIT_CODE=$?

echo "--- Container Output End --- (Exit Code: $EXIT_CODE)"

# --- Post-Execution Processing --- (Moved down here)
if [ $EXIT_CODE -eq 0 ]; then
  echo "Apply test finished successfully."
  echo "Check '${PRIVATE_DIR}' for terraform state."
  echo "Check '${PUBLIC_DIR}' for the generated private key."

  # The container entrypoint now handles outputting lemc directives
  # and saving the key. This script only needs to report success.
  # echo "lemc.html.append;<h4>Saving Private Key...</h4>"
  # Check if key file exists before attempting output commands
  if [ -f "$PRIVATE_KEY_PATH" ]; then
      echo "INFO: Private key found at ${PRIVATE_KEY_PATH}"
      # It's better to let the container output Terraform results via lemc verbs
      # This is just a fallback/local test confirmation
      echo "INFO: (Container should have displayed LB IP and DNS name)"
      echo "INFO: (Container should have provided a download link for the key)"
      echo "INFO: SSH Command hint (using local path and assuming LB IP is routable):"
      echo "       ssh -i ${PRIVATE_KEY_PATH} ${LEMC_USERNAME}@<LOAD_BALANCER_IP>"
  else
      echo "WARNING: Expected private key file was not found at ${PRIVATE_KEY_PATH}. Check container logs."
  fi
else
  echo "Apply test finished with errors (Exit Code: $EXIT_CODE)."
  # Exit the test script with the container's non-zero exit code
  exit $EXIT_CODE
fi

# Unset exported variables if needed (though script exits)
set +a
