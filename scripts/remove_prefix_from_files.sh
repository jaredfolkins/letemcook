#!/bin/bash

# Set the prefix to remove
prefix="embedded_assets_heckle_public_"

# Iterate over all files in the current directory
for old_name in "$prefix"*; do
    # Check if it's actually a file and the name matches the prefix structure
    if [ -f "$old_name" ]; then
        # Construct the new name by removing the prefix
        new_name="${old_name#$prefix}"

        # Check if a file with the new name already exists to avoid overwriting
        if [ -e "$new_name" ]; then
            echo "Skipping rename for '$old_name': '$new_name' already exists."
        else
            # Rename the file
            echo "Renaming '$old_name' to '$new_name'"
            mv -- "$old_name" "$new_name"
        fi
    fi
done

echo "Renaming process complete." 