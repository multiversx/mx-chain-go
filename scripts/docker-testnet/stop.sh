#!/usr/bin/env bash

set -eux

# Delete the entire testnet folder, which includes configuration, executables and logs.

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

# Get the IDs of containers attached to the network
export CONTAINER_IDS=$(docker network inspect -f '{{range $k, $v := .Containers}}{{printf "%s\n" $k}}{{end}}' "$DOCKER_NETWORK_NAME")

mkdir -p ./tmp

# Stop each container
echo "Stopping containers..."
for CONTAINER_ID in $CONTAINER_IDS; do
    docker stop "$CONTAINER_ID"
    echo "$CONTAINER_ID" >> ./tmp/stopped_containers
done