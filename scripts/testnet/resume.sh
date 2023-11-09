#!/usr/bin/env bash

# Resume the paused testnet, by sending SIGCONT to all the processes of the
# testnet (seednode, observers, validators, proxy, txgen)

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$1" == "keep" ]; then
  KEEPOPEN=1
else
  KEEPOPEN=0
fi

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/validators.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/observers.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/tools.sh"

if [ $USE_PROXY -eq 1 ]; then
  resumeProxy
fi

if [ $USE_TXGEN -eq 1 ]; then
  resumeTxGen
fi

resumeValidators
resumeObservers
resumeSeednode
