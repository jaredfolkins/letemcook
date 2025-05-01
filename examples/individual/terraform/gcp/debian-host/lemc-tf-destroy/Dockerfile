# Use an official HashiCorp image with Terraform installed
FROM hashicorp/terraform:latest

# Install jq for JSON processing in entrypoint script
RUN apk add --no-cache jq

# Create a directory for staging Terraform config files
RUN mkdir -p /app/tf_config

# Copy Terraform configuration files from the shared directory
# Context is now the parent directory
COPY terraform-config/*.tf /app/tf_config/

# Copy the entrypoint script into the container (relative to context)
COPY lemc-tf-destroy/entrypoint.sh /usr/local/bin/entrypoint.sh

# Make the entrypoint script executable
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set the entrypoint script as the command to run when the container starts
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Default CMD (hardcoded to destroy for this container)
CMD ["destroy"] 