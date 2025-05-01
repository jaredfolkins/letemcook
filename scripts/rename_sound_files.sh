#!/bin/bash

# Iterate over all items in the current directory
for file in *.mp3; do
  # Check if it's a regular file and not this script itself
  if [ -f "$file" ] && [ "$file" != "$(basename "$0")" ]; then
    # Get the base name and extension
    base_name="${file%.*}"
    extension="${file##*.}"

    # Generate the new base name
    # 1. Convert to lowercase
    # 2. Replace non-alphanumeric characters (allowing underscore) with underscore
    # 3. Squeeze multiple consecutive underscores into one
    # 4. Remove leading underscore
    # 5. Remove trailing underscore
    new_base_name=$(echo "$base_name" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9_' '_' | sed 's/__*/_/g' | sed 's/^_//' | sed 's/_$//')

    # Ensure the new base name is not empty after transformations
    if [ -z "$new_base_name" ]; then
        echo "Warning: Skipping '$file' as its base name resulted in an empty string after transformation."
        continue
    fi

    # Reconstruct the new filename with the original extension
    # Handle cases where the original file might not have an extension
    if [ -n "$extension" ]; then
        new_name="${new_base_name}.${extension}"
    else
        new_name="${new_base_name}"
    fi

    # Check if the new name is different from the old name
    if [ "$file" != "$new_name" ]; then
      # Check if the target filename already exists
      if [ -e "$new_name" ]; then
        echo "Warning: Skipping rename for '$file': Target '$new_name' already exists."
      else
        echo "Renaming '$file' to '$new_name'"
        # Use -- to handle filenames that might start with a hyphen
        mv -- "$file" "$new_name"
      fi
    fi
  fi
done

echo "File renaming process complete."

# Make the script executable
chmod +x "$(basename "$0")" > /dev/null 2>&1 || true # Attempt to make executable, ignore errors if it fails (e.g. no permissions) 