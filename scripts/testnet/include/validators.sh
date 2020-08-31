source "$ELRONDTESTNETSCRIPTSDIR/include/terminal.sh"

startValidators() {
  setTerminalSession "elrond-nodes"
  setTerminalLayout "tiled"
  setWorkdirForNextCommands "$TESTNETDIR/node"
  iterateOverValidators startSingleValidator
}

pauseValidators() {
  iterateOverValidators pauseSingleValidator
}

resumeValidators() {
  iterateOverValidators resumeSingleValidator
}

stopValidators() {
  iterateOverValidators stopSingleValidator
}

iterateOverValidators() {
  local callback=$1
  local VALIDATOR_INDEX=0

  # Iterate over Shard Validators
  (( max_shard_id=$SHARDCOUNT - 1 ))
  for SHARD in $(seq 0 1 $max_shard_id); do
    for _ in $(seq $SHARD_VALIDATORCOUNT); do
      if [ $VALIDATOR_INDEX -ne $SKIP_VALIDATOR_IDX ]; then
        $callback $SHARD $VALIDATOR_INDEX
        sleep 0.5
      fi
      (( VALIDATOR_INDEX++ ))
    done
  done

  # Iterate over Metachain Validators
  SHARD="metachain"
  for _ in $(seq $META_VALIDATORCOUNT); do
    if [ $VALIDATOR_INDEX -ne $SKIP_VALIDATOR_IDX ]; then
      $callback $SHARD $VALIDATOR_INDEX
      sleep 0.5
    fi
     (( VALIDATOR_INDEX++ ))
  done
}

startSingleValidator() {
  local SHARD=$1
  local VALIDATOR_INDEX=$2
  local startCommand="$(assembleCommand_startValidatorNode $VALIDATOR_INDEX)"
  runCommandInTerminal "$startCommand"
}

pauseSingleValidator() {
  local SHARD=$1
  local VALIDATOR_INDEX=$2
  (( PORT=$PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX ))
  pauseProcessByPort $PORT
}

resumeSingleValidator() {
  local SHARD=$1
  local VALIDATOR_INDEX=$2
  (( PORT=$PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX ))
  resumeProcessByPort $PORT
}

stopSingleValidator() {
  local SHARD=$1
  local VALIDATOR_INDEX=$2

  if [ "$NODE_WATCHER" == "1" ]; then
    WORKING_DIR=$TESTNETDIR/node_working_dirs/validator$VALIDATOR_INDEX
    mkdir -p $WORKING_DIR
    touch $WORKING_DIR/norestart
  fi

  (( PORT=$PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX ))
  stopProcessByPort $PORT
}


function isNodeRunning {
  local VALIDATOR_INDEX=$1
  (( PORT=$PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX ))
	local PID=$(lsof -t -i:$PORT)
	if [ -n "$PID" ]; then
		echo 1
	else
		echo 0
	fi
}

function startSingleValidatorWithWatcher {
  local SHARD=$1
  local VALIDATOR_INDEX=$2
  startWatcher $SHARD $VALIDATOR_INDEX &
}

function startWatcher {
  local SHARD=$1
  local VALIDATOR_INDEX=$2
  WORKING_DIR=$TESTNETDIR/node_working_dirs/validator$VALIDATOR_INDEX
  mkdir -p $WORKING_DIR


  if [ -f "$WORKING_DIR/norestart" ]; then
    rm $WORKING_DIR/norestart
  fi

  echo "Starting watcher for validator $VALIDATOR_INDEX..."
  echo "[$(date)] Watcher for validator $VALIDATOR_INDEX started" > $WORKING_DIR/watcher-status
	local probedelay=5
	while true; do
		local running=$(isNodeRunning $VALIDATOR_INDEX)
		if [ "$running" == "0" ]; then
			if [ -f "$WORKING_DIR/norestart" ]; then
        echo "[$(date)] Watcher instructed to stop completely" >> $WORKING_DIR/watcher-status
				break
			fi
			sleep 10
      startSingleValidator $SHARD $VALIDATOR_INDEX
      echo "[$(date)] Validator $VALIDATOR_INDEX found not to be running; started by watcher" >> $WORKING_DIR/watcher-status
			probedelay=60
		else
			probedelay=5
		fi
		sleep $probedelay
	done
  echo "[$(date)] Watcher for validator $VALIDATOR_INDEX exiting..." >> $WORKING_DIR/watcher-status
}

assembleCommand_startValidatorNode() {
  VALIDATOR_INDEX=$1
  (( PORT=$PORT_ORIGIN_VALIDATOR + $VALIDATOR_INDEX ))
  (( RESTAPIPOR=$PORT_ORIGIN_VALIDATOR_REST + $VALIDATOR_INDEX ))
  (( KEY_INDEX=$VALIDATOR_INDEX ))
  WORKING_DIR=$TESTNETDIR/node_working_dirs/validator$VALIDATOR_INDEX

  local nodeCommand="./node \
        -port $PORT --profile-mode -log-save -log-level $LOGLEVEL --log-logger-name --log-correlation --use-health-service -rest-api-interface localhost:$RESTAPIPORT \
        -sk-index $KEY_INDEX \
        -working-directory $WORKING_DIR -config ./config/config_validator.toml"

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
