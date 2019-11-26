#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

DEPLOYED_SC_ADDRESS=$(cat $TESTNETDIR/txgen/deployedSCAddress.txt)

curl -d '{
  "value": 10,
  "numOfTxs": 5,
  "numOfShards": '$SHARDCOUNT',
  "gasPrice": 1,
  "gasLimit": 5000000,
  "crossShard": true,
  "recallNonce": true,
  "scAddress": "'$DEPLOYED_SC_ADDRESS'",
  "data": "transfer"
}' \
  -H "Content-Type: application/json" -X POST http://localhost:$PORT_TXGEN/transaction/send-multiple
