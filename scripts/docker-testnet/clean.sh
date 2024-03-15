#!/usr/bin/env bash

# Delete the entire testnet folder, which includes configuration, executables and logs.

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

echo "Stopping all containers..."
docker stop $(docker ps -a -q)

echo "Removing all containers..."
docker container prune -f

echo "Removing network..."
docker network rm ${DOCKER_NETWORK_NAME}

echo "Removing $TESTNETDIR..."
rm -rf $TESTNETDIR
