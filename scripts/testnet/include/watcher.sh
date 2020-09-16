function node-start-with-watcher {
  local VALIDATOR_INDEX=$1
  local PORT=$2
  local WORKING_DIR="$(pwd)"
  local TESTNETDIR="../.."
  
  if [ -f "$WORKING_DIR/norestart" ]; then
    rm $WORKING_DIR/norestart
  fi

  echo "Starting watcher for validator $VALIDATOR_INDEX..."
  echo "[$(date)] Watcher for validator $VALIDATOR_INDEX started" > $WORKING_DIR/watcher-status
  local node_command=$(cat $WORKING_DIR/node-command)

	local probedelay=5
	while true; do
		local running=$(isNodeRunning $PORT)
		if [ "$running" == "0" ]; then
			if [ -f "$WORKING_DIR/norestart" ]; then
        echo "[$(date)] Watcher instructed to stop completely" >> $WORKING_DIR/watcher-status
				break
			fi
			sleep 10

      echo "[$(date)] Validator $VALIDATOR_INDEX found not to be running" >> $WORKING_DIR/watcher-status
      echo "[$(date)] Starting node:\n$node_command" >> $WORKING_DIR/watcher-status

      pushd $TESTNETDIR/node
      pwd
      $node_command &
      popd

			probedelay=60
		else
			probedelay=5
		fi
		sleep $probedelay
	done
  echo "[$(date)] Watcher for validator $VALIDATOR_INDEX exiting..." >> $WORKING_DIR/watcher-status
}

function isNodeRunning {
  local PORT=$1
	local PID=$(lsof -t -i:$PORT)
	if [ -n "$PID" ]; then
		echo 1
	else
		echo 0
	fi
}
