#!/usr/bin/env bash

set -eux

export DOCKERTESTNETDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

MULTIVERSXTESTNETSCRIPTSDIR="$(dirname "$DOCKERTESTNETDIR")/testnet"

source "$DOCKERTESTNETDIR/variables.sh"

cd ${MULTIVERSXDIR}

docker build -f docker/seednode/Dockerfile . -t seednode:dev

docker build -f docker/node/Dockerfile . -t node:dev

