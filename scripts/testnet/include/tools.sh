source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startProxy() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-vertical"

  setWorkdirForNextCommands "$TESTNETDIR/proxy"
  runCommandInTerminal "./proxy" $1
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

  runCommandInTerminal "./txgen -num-accounts $NUMACCOUNTS -new-accounts |& tee stdout.txt" $1
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

startSeednode() {
  setTerminalSession "elrond-tools"
  setTerminalLayout "even-horizontal"

  setWorkdirForNextCommands "$TESTNETDIR/seednode"

  if [ -n "$NODE_NICENESS" ]
  then
    seednodeCommand="nice -n $NODE_NICENESS ./seednode"
  else
    seednodeCommand="./seednode"
  fi

  runCommandInTerminal "$seednodeCommand" $1
}

pauseSeednode() {
  pauseProcessByPort $PORT_SEEDNODE
}

resumeSeednode() {
  resumeProcessByPort $PORT_SEEDNODE
}

stopSeednode() {
  stopProcessByPort $PORT_SEEDNODE
}
