#!/bin/sh

# Create a file in the public mount
# Ensure the directory exists (it should, but defensive programming)
mkdir -p /lemc/public
echo "Hello from the public mount, created at $(date)!" > /lemc/public/hello_public.txt

# Function to append raw HTML content
lemc_html_buffer() {
    # Note: Caller must ensure $1 is valid HTML snippet
    echo "lemc.html.buffer; $1"
}

# Function to append raw CSS content
lemc_css_buffer() {
    # Note: Caller must ensure $1 is valid CSS
    echo "lemc.css.buffer; $1"
}

# Function to truncate HTML & CSS content areas
lemc_truncate() {
    echo "lemc.html.trunc;"
    echo "lemc.css.trunc;"
}

# Function to send CSS file content line by line
send_css() {
    local css_file="$1"
    local target_id="${LEMC_HTML_ID:-lemc-output-container}" # Use env var or default

    if [ ! -f "$css_file" ]; then
        echo "lemc.error; CSS file not found: $css_file" >&2
        return 1
    fi

    # Process with sed and pipe line by line
    sed "s/%%target_id%%/$target_id/g" "$css_file" | while IFS= read -r line
do
    lemc_css_buffer "$line"
done
}

# Function to send HTML template content line by line, replacing placeholders
send_html_template() {
    local html_file="$1"

    if [ ! -f "$html_file" ]; then
        echo "lemc.error; HTML template file not found: $html_file" >&2
        return 1
    fi

    # Prepare values (handle unset variables gracefully)
    local date_val=$(date)
    local uuid_val=${LEMC_UUID:-"Not Set"}
    local scope_val=${LEMC_SCOPE:-"Not Set"}
    local userid_val=${LEMC_USER_ID:-"Not Set"}
    local username_val=${LEMC_USERNAME:-"Not Set"}
    local recipe_name_val=${LEMC_RECIPE_NAME:-"Not Set"}
    local page_id_val=${LEMC_PAGE_ID:-"Not Set"}
    local step_id_val=${LEMC_STEP_ID:-"Not Set"}
    local http_dl_val=${LEMC_HTTP_DOWNLOAD_BASE_URL:-"Not Set"}
    local public_var_val=${USER_DEFINED_PUBLIC_ENV_VAR:-"Not Set"}
    local private_var_val=${USER_DEFINED_PRIVATE_ENV_VAR:-"Not Set"} # Note: Example only
    local final_msg_val="Time to cook baby!"

    # Construct the download link
    local download_filename="hello_public.txt"
    local download_url="${http_dl_val}${download_filename}"
    local download_link="<a href=\"${download_url}\" download=\"${download_filename}\">Download ${download_filename}</a>"
    # Escape for sed: replace / with \\/ and & with \\&
    local escaped_download_link=$(echo "$download_link" | sed -e 's/[\/&]/\\&/g')

    # Process with sed and pipe line by line for replacement
    sed \
      -e "s|%%date%%|$date_val|g" \
      -e "s|%%LEMC_UUID%%|$uuid_val|g" \
      -e "s|%%LEMC_SCOPE%%|$scope_val|g" \
      -e "s|%%LEMC_USER_ID%%|$userid_val|g" \
      -e "s|%%LEMC_USERNAME%%|$username_val|g" \
      -e "s|%%LEMC_RECIPE_NAME%%|$recipe_name_val|g" \
      -e "s|%%LEMC_PAGE_ID%%|$page_id_val|g" \
      -e "s|%%LEMC_STEP_ID%%|$step_id_val|g" \
      -e "s|%%LEMC_HTTP_DOWNLOAD_BASE_URL%%|$http_dl_val|g" \
      -e "s|%%USER_DEFINED_PUBLIC_ENV_VAR%%|$public_var_val|g" \
      -e "s|%%USER_DEFINED_PRIVATE_ENV_VAR%%|$private_var_val|g" \
      -e "s|%%DOWNLOAD_LINK_PLACEHOLDER%%|$escaped_download_link|g" \
      -e "s|%%final_message%%|$final_msg_val|g" \
      "$html_file" | while IFS= read -r line
do
    lemc_html_buffer "$line"
done
}

# --- Main Script --- 

# Truncate CSS & HTML area
lemc_truncate

# Send CSS
send_css "/styles.css"

# Send processed HTML template
send_html_template "/template.html" 

echo "lemc.css.append;"
echo "lemc.html.append;"