#!/usr/bin/env bash

set -eux

cd ${MULTIVERSXDIR}

docker build -f docker/seednode/Dockerfile . -t seednode:dev

docker build -f docker/node/Dockerfile . -t node:dev

