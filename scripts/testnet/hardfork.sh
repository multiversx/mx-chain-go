#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/config.sh"

VALIDATOR_RES_PORT="$PORT_ORIGIN_VALIDATOR_REST"

if [ -z "$1" ]; then
  echo "epoch argument was not provided. Usage: './hardfork.sh [epoch number] [forced: true|false]' as in './hardfork.sh 1 false'"
  exit
fi

if [ -z "$2" ]; then
  echo "epoch argument was not provided. Usage: './hardfork.sh [epoch: number] [forced: true|false]' as in './hardfork.sh 1 false'"
  exit
fi

if [ $1 -lt "1" ]; then
  echo "incorrect epoch argument was provided. Usage: './hardfork.sh [epoch number]' as in './hardfork.sh 1'"
  exit
fi

address="http://127.0.0.1:$VALIDATOR_RES_PORT/hardfork/trigger"
epoch=$1
forced=$2
curl -d '{"epoch":'"$epoch"',"forced":'"$forced"'}' -H 'Content-Type: application/json' $address

echo " done curl"

# change the setting from config.toml: AfterHardFork to true
updateTOMLValue "$TESTNETDIR/node/config/config_validator.toml" "AfterHardFork" "true"
updateTOMLValue "$TESTNETDIR/node/config/config_observer.toml" "AfterHardFork" "true"

# change nodesSetup.json genesis time to a new value
let startTime="$(date +%s) + $HARDFORK_DELAY"
updateJSONValue "$TESTNETDIR/node/config/nodesSetup.json" "startTime" "$startTime"

updateTOMLValue "$TESTNETDIR/node/config/config_validator.toml" "GenesisTime" $startTime
updateTOMLValue "$TESTNETDIR/node/config/config_observer.toml" "GenesisTime" $startTime

# copy back the configs
if [ $COPY_BACK_CONFIGS -eq 1 ]; then
  copyBackConfigs
fi

echo "done hardfork reconfig"
