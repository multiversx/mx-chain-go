source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startProxy() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/proxy"
  runCommandInTerminal "./proxy" $1 h
}

stopProxy() {
  stopProcessByPort $PORT_PROXY
}

startTxGen_NewAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" h
  runCommandInTerminal "./txgen -num-accounts $NUMACCOUNTS -num-shards $SHARDCOUNT -new-accounts" $1
}

startTxGen_ExistingAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" h
  runCommandInTerminal "./txgen" $1
}

stopTxGen() {
  stopProcessByPort $PORT_TXGEN
}
