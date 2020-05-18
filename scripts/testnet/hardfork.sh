#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

# Load local overrides, .gitignored
LOCAL_OVERRIDES="$ELRONDTESTNETSCRIPTSDIR/local.sh"
if [ -f "$LOCAL_OVERRIDES" ]; then
  source "$ELRONDTESTNETSCRIPTSDIR/local.sh"
fi

VALIDATOR_RES_PORT="$PORT_ORIGIN_VALIDATOR_REST"

if [ -z "$1" ]; then
  echo "epoch argument was not provided. Usage: './hardfork.sh [epoch number]' as in './hardfork.sh 0'"
  exit
fi

epoch=$1
cmd=(printf "$(curl -d '{"epoch":'"$epoch"'}' -H 'Content-Type: application/json' http://127.0.0.1:$VALIDATOR_RES_PORT/hardfork/trigger)")
"${cmd[@]}"

echo " done curl"
