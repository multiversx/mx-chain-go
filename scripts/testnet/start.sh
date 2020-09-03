#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export TESTNETMODE=$1
export EXTRA=$2

source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/config.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/build.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/validators.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/observers.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/tools.sh"

prepareFolders

# Phase 1: build Seednode, Node and Arwen executables
buildSeednode
buildNode

if [ $ALWAYS_BUILD_ARWEN -eq 1 ]; then
  buildArwen
fi


# Phase 2: generate configuration
if [ $ALWAYS_UPDATE_CONFIGS -eq 1 ]; then
  copyConfig
fi

copySeednodeConfig
updateSeednodeConfig

if [ $ALWAYS_UPDATE_CONFIGS -eq 1 ]; then
  copyNodeConfig
  updateNodeConfig
fi


# Phase 3: start the Seednode
startSeednode
showTerminalSession "elrond-tools"
echo "Waiting for the Seednode to start ($SEEDNODE_DELAY s)..."
sleep $SEEDNODE_DELAY

# Phase 4: start the Observer Nodes and Validator Nodes
startObservers
startValidators
showTerminalSession "elrond-nodes"
echo "Waiting for the Nodes to start ($NODE_DELAY s)..."
sleep $NODE_DELAY

# Phase 5: build the Proxy and TxGen
if [ $USE_PROXY -eq 1 ]; then
  prepareFolders_Proxy
  buildProxy
fi
if [ $USE_TXGEN -eq 1 ]; then
  prepareFolders_TxGen
  buildTxGen
fi

# Phase 6: start the Proxy
if [ $USE_PROXY -eq 1 ]; then
  startProxy
  echo "Waiting for the Proxy to start ($PROXY_DELAY s)..."
  sleep $PROXY_DELAY
fi

# Phase 7: start the TxGen, with or without regenerating the accounts
if [ $USE_TXGEN -eq 1 ]; then
  if [ -n "$TXGEN_REGENERATE_ACCOUNTS" ]
  then
    echo "Starting TxGen with account generation..."
    startTxGen_NewAccounts
  else
    echo "Starting TxGen with existing accounts..."
    startTxGen_ExistingAccounts
  fi
fi
