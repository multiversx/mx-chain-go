#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

DEPLOYED_SC_ADDRESS=$(cat $TESTNETDIR/txgen/deployedSCAddress.txt)


sendTransactions() {
  curl -d '{
    "value": 10,
    "numOfTxs": 200,
    "numOfShards": '$SHARDCOUNT',
    "gasPrice": 1,
    "gasLimit": 3000000,
    "crossShard": true,
    "recallNonce": true,
    "scAddress": "'$DEPLOYED_SC_ADDRESS'",
    "data": "transfer"
  }' \
    -H "Content-Type: application/json" -X POST http://localhost:$PORT_TXGEN/transaction/send-multiple
}

while true
do
  sendTransactions
  echo "Transaction batch emitted"
  sleep 600

  validation=$( curl -s "http://127.0.0.1:$PORT_TXGEN/validate/sc/$DEPLOYED_SC_ADDRESS" | jq -s | grep ': {' | wc -l )
  echo "Validation result: $validation errors"
done
