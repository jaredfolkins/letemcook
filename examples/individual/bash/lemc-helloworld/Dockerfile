FROM alpine

# Copy the entrypoint script and assets
COPY entrypoint.sh /entrypoint.sh
COPY styles.css /styles.css
COPY template.html /template.html

# Make entrypoint executable
RUN chmod +x /entrypoint.sh

# Set the entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Original CMD removed
# CMD ["/bin/bash", "-c", "echo \"lemc.html.trunc; Hello, World!\""]

