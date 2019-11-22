#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

DEPLOYED_SC_ADDRESS=$(cat $TESTNETDIR/txgen/deployedSCAddress.txt)

curl -d '{
  "value": 10,
  "numOfTxs": 2000,
  "numOfShards": '$SHARDCOUNT',
  "gasPrice": 10,
  "gasLimit": 1000000000,
  "crossShard": true,
  "recallNonce": true,
  "scAddress": "$DEPLOYED_SC_ADDRESS",
  "data": "transfer"
}' \
  -H "Content-Type: application/json" -X POST http://localhost:$PORT_TXGEN/transaction/send-multiple
