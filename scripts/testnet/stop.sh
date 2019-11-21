#!/bin/env sh

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$ELRONDTESTNETSCRIPTSDIR/include/nodes.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/tools.sh"

stopTxGen
stopProxy
stopValidators
stopObservers
stopSeednode
