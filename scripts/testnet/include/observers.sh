source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startObservers() {
  setTerminalSession "elrond-nodes"
  setTerminalLayout "tiled"
  setWorkdirForNextCommands "$TESTNETDIR/node"
  iterateOverObservers startSingleObserver
}

pauseObservers() {
  iterateOverObservers pauseSingleObserver
}

resumeObservers() {
  iterateOverObservers resumeSingleObserver
}

stopObservers() {
  iterateOverObservers stopSingleObserver
}

iterateOverObservers() {
  local callback=$1
  local OBSERVER_INDEX=0

  # Iterate over Shard Observers
  (( max_shard_id=$SHARDCOUNT - 1 ))
  for SHARD in $(seq 0 1 $max_shard_id); do
    for _ in $(seq $SHARD_OBSERVERCOUNT); do
      if [ $OBSERVER_INDEX -ne $SKIP_OBSERVER_IDX ]; then
        $callback $SHARD $OBSERVER_INDEX
        sleep 0.2
      fi
      (( OBSERVER_INDEX++ ))
    done
  done

  # Iterate over Metachain Observers
  local SHARD="metachain"
  for META_OBSERVER in $(seq $META_OBSERVERCOUNT); do
    if [ $OBSERVER_INDEX -ne $SKIP_OBSERVER_IDX ]; then
      $callback $SHARD $OBSERVER_INDEX
      sleep 0.2
    fi
    (( OBSERVER_INDEX++ ))
  done
}

startSingleObserver() {
  local SHARD=$1
  local OBSERVER_INDEX=$2
  local startCommand="$(assembleCommand_startObserverNode $SHARD $OBSERVER_INDEX)"
  runCommandInTerminal "$startCommand"
}

pauseSingleObserver() {
  local SHARD=$1
  local OBSERVER_INDEX=$2
  let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
  pauseProcessByPort $PORT
}

resumeSingleObserver() {
  local SHARD=$1
  local OBSERVER_INDEX=$2
  let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
  resumeProcessByPort $PORT
}

stopSingleObserver() {
  local SHARD=$1
  local OBSERVER_INDEX=$2
  let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
  stopProcessByPort $PORT
}

assembleCommand_startObserverNode() {
  SHARD=$1
  OBSERVER_INDEX=$2
  let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
  let "RESTAPIPORT=$PORT_ORIGIN_OBSERVER_REST + $OBSERVER_INDEX"
  let "KEY_INDEX=$TOTAL_NODECOUNT - $OBSERVER_INDEX - 1"
  WORKING_DIR=$TESTNETDIR/node_working_dirs/observer$OBSERVER_INDEX

  local nodeCommand="./node \
        -port $PORT --profile-mode -log-save -log-level $LOGLEVEL --log-logger-name --log-correlation --use-health-service -rest-api-interface localhost:$RESTAPIPORT \
        -destination-shard-as-observer $SHARD \
        -sk-index $KEY_INDEX \
        -working-directory $WORKING_DIR -config ./config/config_observer.toml"

  if [ -n "$NODE_NICENESS" ]
  then
    nodeCommand="nice -n $NODE_NICENESS $nodeCommand"
  fi

  if [ $NODETERMUI -eq 0 ]
  then
    nodeCommand="$nodeCommand -use-log-view"
  fi

  echo $nodeCommand
}
