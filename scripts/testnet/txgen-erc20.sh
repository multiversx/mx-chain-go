#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

DESTINATION=mixed
if [[ -n $1 ]]; then
  DESTINATION=$1
fi

NUM_SHARDS=${SHARDCOUNT}
NUM_TXS=25
VALUE=10
GAS_PRICE=200000000000
GAS_LIMIT=3000000
SLEEP_DURATION=40
DEPLOYED_SC_ADDRESS=$(cat $TESTNETDIR/txgen/deployedSCAddress.txt)

echo 'Starting ERC20 scenario cURL script...'
echo 'Will run a cURL with the following JSON payload each '$SLEEP_DURATION' seconds'
echo '{"value": '$VALUE', "numOfTxs": '$NUM_TXS', "numOfShards": '$NUM_SHARDS', "gasPrice": '$GAS_PRICE', "gasLimit": '$GAS_LIMIT', "destination": "'$DESTINATION'", "recallNonce": true, "scenario": "erc20", "scAddress": "'$DEPLOYED_SC_ADDRESS'", "data": "transfer"}'
echo
echo 'Starting...'

for (( ; ; )); do
  curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "numOfShards": '$NUM_SHARDS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "scenario": "erc20",
     "scAddress": "'$DEPLOYED_SC_ADDRESS'",
     "data": "transfer"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple

  sleep $SLEEP_DURATION
done
