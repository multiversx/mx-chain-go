#!/usr/bin/env bash

set -eux

export DOCKERTESTNETDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

file_path="${DOCKERTESTNETDIR}/tmp/stopped_containers"

# Check if the file exists
if [ ! -f "$file_path" ]; then
    echo "File $file_path not found."
    exit 1
fi

# Read the file line by line
while IFS= read -r line; do
    docker start $line
done < "$file_path"

rm -rf "${DOCKERTESTNETDIR}/tmp"