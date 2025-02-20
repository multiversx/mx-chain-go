#!/usr/bin/env bash

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
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/build.sh"

if [ $USE_PROXY -eq 1 ]; then
  stopProxy
fi

if [ $USE_TXGEN -eq 1 ]; then
  stopTxGen
fi

stopValidators
stopObservers
stopSeednode

if [ $USE_ELASTICSEARCH -eq 1 ]; then
  stopElasticsearch
fi

if [ $USETMUX -eq 1 ] && [ $KEEPOPEN -eq 0 ]
then
  tmux kill-session -t "multiversx-tools"
  tmux kill-session -t "multiversx-nodes"
fi
