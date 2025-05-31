#!/bin/bash

# Set the target directory (relative to this script's location)
TARGET_DIR="../../public"

echo "Starting file renaming and moving process..."
echo "Processing files in: $(pwd)"
echo "Moving renamed files to: $TARGET_DIR"

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
  
  echo "Processing: '$file'"
  
  # If the name would change, rename it first
  if [ "$file" != "$newname" ]; then
    echo "  Renaming to: '$newname'"
    mv "$file" "$newname"
    file="$newname"  # Update the variable for the move operation
  else
    echo "  No renaming needed"
  fi
  
  # Move the file to the target directory (will replace if exists)
  echo "  Moving to: $TARGET_DIR/$file"
  mv "$file" "$TARGET_DIR/"
  
done

echo "File renaming and moving complete!"
echo "All files have been moved from normalized-all to public directory." 