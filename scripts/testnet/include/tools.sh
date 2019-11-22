source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startProxy() {
  startTerminal

  cd $TESTNETDIR/proxy
  runCommandInTerminal "./proxy" $1

  endTerminal
}

stopProxy() {
  stopProcessByPort $PORT_PROXY
}

startTxGen_NewAccounts() {
  startTerminal

  cd $TESTNETDIR/txgen
  runCommandInTerminal "./txgen -num-accounts $NUMACCOUNTS -num-shards $SHARDCOUNT -new-accounts -sc-mode" $1

  endTerminal
}

startTxGen_ExistingAccounts() {
  startTerminal

  cd $TESTNETDIR/txgen
  runCommandInTerminal "./txgen" $1

  endTerminal
}

stopTxGen() {
  stopProcessByPort $PORT_TXGEN
}
