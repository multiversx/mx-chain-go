#!/usr/bin/env bash

set -e

export DOCKERTESTNETDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

MULTIVERSXTESTNETSCRIPTSDIR="$(dirname "$DOCKERTESTNETDIR")/testnet"

source "$DOCKERTESTNETDIR/variables.sh"
source "$DOCKERTESTNETDIR/functions.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/config.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/build.sh"

prepareFolders

buildConfigGenerator

generateConfig

copyConfig

copySeednodeConfig
updateSeednodeConfig

copyNodeConfig
updateNodeConfig

startSeedNode
startObservers
startValidators

if [ $USE_PROXY -eq 1 ]; then
  prepareFolders_Proxy
  copyProxyConfig
  updateProxyConfigDocker
  startProxyDocker
fi

