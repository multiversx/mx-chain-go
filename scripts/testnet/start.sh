#!/bin/env sh

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/config.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/build.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/nodes.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/tools.sh"

prepareFolders

# Phase 1: build Seednode and Node executables
buildSeednode
buildNode
buildProxy
buildTxGen

# Phase 2: generate configuration
generateConfig
copyConfig
copySeednodeConfig
copyNodeConfig

copyProxyConfig
updateProxyConfig
copyTxGenConfig
updateTxGenConfig

# Phase 3: start the Seednode
startSeednode
sleep $SEEDNODE_DELAY

# Phase 4: start the Observer Nodes and Validator Nodes
startObservers
startValidators
sleep $NODE_DELAY

# Phase 5: start the Proxy
startProxy
sleep $PROXY_DELAY

# Phase 6: start the TxGen, with or without regenerating the accounts
if [ -n "$REGENERATE_ACCOUNTS" ]
then
  startTxGen_NewAccounts
else
  startTxGen_ExistingAccounts
fi
