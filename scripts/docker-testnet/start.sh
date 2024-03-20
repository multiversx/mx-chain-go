#!/usr/bin/env bash

set -eux

file_path="./tmp/stopped_containers"

# Check if the file exists
if [ ! -f "$file_path" ]; then
    echo "File $file_path not found."
    exit 1
fi

# Read the file line by line
while IFS= read -r line; do
    docker start $line
done < "$file_path"

rmdir ./tmp