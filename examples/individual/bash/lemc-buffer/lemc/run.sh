#!/bin/bash

echo "running lemc-buffer"

timestamp=$(date +%s)
filepath="/lemc/public/output.txt"
tmp="/lemc/public/tmp.txt"

# append timestamp to the file
echo "$timestamp<br>" >> $filepath

# Initialize the counter
line_count=$(wc -l < "$filepath")

# Check if the total line count is greater than 50
if [ "$line_count" -gt 10 ]; then
  # Delete the first line and save the result back to the file
  tail -n +2 "$filepath" > $tmp
  mv $tmp $filepath
fi

# Read the file content and replace newlines with spaces
file_content=$(tr '\n' ' ' < "$filepath")

# Trim trailing spaces
file_content=$(echo "$file_content" | sed 's/[[:space:]]*$//')

echo "lemc.html.trunc; $file_content"