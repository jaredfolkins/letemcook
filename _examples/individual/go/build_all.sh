#!/bin/bash

# Iterate through all directories in the current directory
for dir in */; do
  # Check if build.sh exists in the directory
  if [ -f "$dir/build.sh" ]; then
    # Run the build.sh script
    (cd "$dir" && ./build.sh)
  fi
done