#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/config.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/build.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/nodes.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/tools.sh"

prepareFolders

# Phase 1: build Seednode and Node executables
buildConfigGenerator
buildSeednode
buildNode
buildProxy
buildTxGen

# Phase 2: generate configuration
generateConfig
copyConfig

copySeednodeConfig
updateSeednodeConfig

copyNodeConfig
updateNodeConfig

copyProxyConfig
updateProxyConfig
copyTxGenConfig
updateTxGenConfig

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

# Phase 5: start the Proxy
startProxy
echo "Waiting for the Proxy to start ($PROXY_DELAY s)..."
sleep $PROXY_DELAY

# Phase 6: start the TxGen, with or without regenerating the accounts
if [ -n "$REGENERATE_ACCOUNTS" ]
then
  echo "Starting TxGen with account generation..."
  startTxGen_NewAccounts
else
  echo "Starting TxGen with existing accounts..."
  startTxGen_ExistingAccounts
fi
