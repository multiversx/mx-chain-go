source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"
source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startSeednode() {
  startTerminal

  cd $TESTNETDIR/seednode
  runCommandInTerminal "./seednode -port $PORT_SEEDNODE" $1

  endTerminal
}

stopSeednode() {
  stopProcessByPort $PORT_SEEDNODE
}

startObservers() {
  startTerminal

  cd $TESTNETDIR/node

  OBSERVER_INDEX=0
  # Start Shard Observers
  let "max_shard_id=$SHARDCOUNT - 1"
  for SHARD in `seq 0 1 $max_shard_id`; do
    for OBSERVER_IN_SHARD in `seq $SHARD_OBSERVERCOUNT`; do

      runCommandInTerminal "$(assembleCommand_startObserverNode $SHARD $OBSERVER_INDEX)" $1

      let OBSERVER_INDEX++
    done
  done

  endTerminal

  startTerminal
  # Start Metachain Observers
  SHARD="metachain"
  for META_OBSERVER in `seq $META_OBSERVERCOUNT`; do

    runCommandInTerminal "$(assembleCommand_startObserverNode $SHARD $OBSERVER_INDEX)" $1

    let OBSERVER_INDEX++
  done

  endTerminal
}

stopObservers() {
  OBSERVER_INDEX=0
  # Stop Shard Observers
  let "max_shard_id=$SHARDCOUNT - 1"
  for SHARD in `seq 0 1 $max_shard_id`; do
    for OBSERVER_IN_SHARD in `seq $SHARD_OBSERVERCOUNT`; do

      let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
      stopProcessByPort $PORT

      let OBSERVER_INDEX++
    done
  done

  endTerminal

  startTerminal
  # Start Metachain Observers
  SHARD="metachain"
  for META_OBSERVER in `seq $META_OBSERVERCOUNT`; do

      let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
      stopProcessByPort $PORT

    let OBSERVER_INDEX++
  done
}

startValidators() {
  startTerminal

  cd $TESTNETDIR/node

  VALIDATOR_INDEX=0
  let "max_shard_id=$SHARDCOUNT - 1"
  for SHARD in `seq 0 1 $max_shard_id`; do
    for VALIDATOR_IN_SHARD in `seq $SHARD_VALIDATORCOUNT`; do

      runCommandInTerminal "$(assembleCommand_startValidatorNode $VALIDATOR_INDEX)" $1

      let VALIDATOR_INDEX++
    done
  done

  endTerminal
}

stopValidators() {
  VALIDATOR_INDEX=0
  let "max_shard_id=$SHARDCOUNT - 1"
  for SHARD in `seq 0 1 $max_shard_id`; do
    for VALIDATOR_IN_SHARD in `seq $SHARD_VALIDATORCOUNT`; do

      let "PORT = $PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX"
      stopProcessByPort $PORT

      let VALIDATOR_INDEX++
    done
  done
}

assembleCommand_startObserverNode() {
  SHARD=$1
  OBSERVER_INDEX=$2
  let "PORT = $PORT_ORIGIN_OBSERVER + $OBSERVER_INDEX"
  let "RESTAPIPORT=$PORT_ORIGIN_OBSERVER_REST + $OBSERVER_INDEX"
  let "KEY_INDEX=$TOTAL_NODECOUNT - $OBSERVER_INDEX - 1"
  WORKING_DIR=$TESTNETDIR/node_working_dirs/observer$OBSERVER_INDEX

  echo "./node \
        -port $PORT -rest-api-interface localhost:$RESTAPIPORT \
        -tx-sign-sk-index $KEY_INDEX -sk-index $KEY_INDEX \
        -num-of-nodes $TOTAL_NODECOUNT -storage-cleanup -destination-shard-as-observer $SHARD \
        -working-directory $WORKING_DIR"
}

assembleCommand_startValidatorNode() {
  VALIDATOR_INDEX=$1
  let "PORT = $PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX"
  let "RESTAPIPORT=$PORT_ORIGIN_VALIDATOR_REST + $VALIDATOR_INDEX"
  let "KEY_INDEX=$VALIDATOR_INDEX"
  WORKING_DIR=$TESTNETDIR/node_working_dirs/validator$VALIDATOR_INDEX

  echo "./node \
        -port $PORT -rest-api-interface localhost:$RESTAPIPORT \
        -tx-sign-sk-index $KEY_INDEX -sk-index $KEY_INDEX \
        -num-of-nodes $TOTAL_NODECOUNT -storage-cleanup \
        -working-directory $WORKING_DIR"
}
