#!/usr/bin/env bash

set -eux

# Delete the entire testnet folder, which includes configuration, executables and logs.

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

# Get the IDs of containers attached to the network
CONTAINER_IDS=$(docker network inspect -f '{{range .Containers}}{{.Name}} {{end}}' "$DOCKER_NETWORK_NAME")

# Stop each container
echo "Removing containers..."
for CONTAINER_ID in $CONTAINER_IDS; do
    docker stop "$CONTAINER_ID"
    docker rm "$CONTAINER_ID"
done

echo "Removing network..."
docker network rm ${DOCKER_NETWORK_NAME}

echo "Removing $TESTNETDIR..."
rm -rf $TESTNETDIR
