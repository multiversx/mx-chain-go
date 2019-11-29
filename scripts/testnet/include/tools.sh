source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startProxy() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/proxy"
  runCommandInTerminal "./proxy" $1 v
}

stopProxy() {
  stopProcessByPort $PORT_PROXY
}

startTxGen_NewAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" v

  local mode=""
  if [ $TXGEN_ERC20_MODE -eq 1 ]; then
    mode="-sc-mode"
    echo "TxGen will start in ERC20 mode"
  fi

  runCommandInTerminal "./txgen -num-accounts $NUMACCOUNTS -num-shards $SHARDCOUNT -new-accounts $mode |& tee stdout.txt" $1
}

startTxGen_ExistingAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" v

  local mode=""
  if [ $TXGEN_ERC20_MODE -eq 1 ]; then
    mode="-sc-mode"
    echo "TxGen will start in ERC20 mode"
  fi

  runCommandInTerminal "./txgen $mode |& tee stdout.txt" $1
}

stopTxGen() {
  stopProcessByPort $PORT_TXGEN
}
