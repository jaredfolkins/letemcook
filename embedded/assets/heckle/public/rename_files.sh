#!/bin/bash

echo "Starting file renaming process..."

for file in *.mp3; do
  # Skip if no files match the pattern
  [ ! -f "$file" ] && continue
  
  # Remove special characters (keep only alphanumeric, spaces, dots, underscores, and hyphens)
  # Then replace spaces with underscores
  # Clean up multiple consecutive underscores
  # Remove leading underscores
  # Fix the extension if needed
  newname=$(echo "$file" | \
    sed 's/[^a-zA-Z0-9 ._-]//g' | \
    sed 's/ /_/g' | \
    sed 's/__*/_/g' | \
    sed 's/^_//g' | \
    sed 's/_\.mp3$/.mp3/g')
  
  # Only rename if the name would actually change
  if [ "$file" != "$newname" ]; then
    echo "Renaming: '$file' -> '$newname'"
    mv "$file" "$newname"
  else
    echo "No change needed: '$file'"
  fi
done

echo "File renaming complete!" 