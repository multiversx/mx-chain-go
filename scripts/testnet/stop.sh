#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$1" == "keep" ]; then
  KEEPOPEN=1
else
  KEEPOPEN=0
fi

source "$ELRONDTESTNETSCRIPTSDIR/include/nodes.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/tools.sh"

if [ $PRIVATE_REPOS -eq 1 ]; then
  stopTxGen
  stopProxy
fi

stopValidators
stopObservers
stopSeednode

if [ $USETMUX -eq 1 ] && [ $KEEPOPEN -eq 0 ]
then
  tmux kill-session -t "elrond-tools"
  tmux kill-session -t "elrond-nodes"
fi
