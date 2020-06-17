source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startProxy() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-vertical"

  setWorkdirForNextCommands "$TESTNETDIR/proxy"
  runCommandInTerminal "./proxy" $1 v
}

pauseProxy() {
  pauseProcessByPort $PORT_PROXY
}

resumeProxy() {
  resumeProcessByPort $PORT_PROXY
}

stopProxy() {
  stopProcessByPort $PORT_PROXY
}

startTxGen_NewAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-vertical"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" v

  runCommandInTerminal "./txgen -num-accounts $NUMACCOUNTS -num-shards $SHARDCOUNT -new-accounts |& tee stdout.txt" $1
}

startTxGen_ExistingAccounts() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-vertical"

  setWorkdirForNextCommands "$TESTNETDIR/txgen" v

  runCommandInTerminal "./txgen |& tee stdout.txt" $1
}

pauseTxGen() {
  pauseProcessByPort $PORT_TXGEN
}

resumeTxGen() {
  resumeProcessByPort $PORT_TXGEN
}

stopTxGen() {
  stopProcessByPort $PORT_TXGEN
}
