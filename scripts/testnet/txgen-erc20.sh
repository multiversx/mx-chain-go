#!/usr/bin/env bash

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

DESTINATION=mixed
if [[ -n $1 ]]; then
  DESTINATION=$1
fi

NUM_TXS=10000
VALUE=1
GAS_PRICE=1000000000
GAS_LIMIT=3000000
SLEEP_DURATION=20
DEPLOY_SLEEP_DURATION=25
MINT_SLEEP_DURATION=25

echo 'Starting ERC20 scenario cURL script...'
echo 'Will run a cURL with the following JSON payload each '$SLEEP_DURATION' seconds'
echo '{"value": '$VALUE', "numOfTxs": '$NUM_TXS', "gasPrice": '$GAS_PRICE', "gasLimit": '$GAS_LIMIT', "destination": "'$DESTINATION'", "recallNonce": true, "scenario": "erc20", "scAddress": "'$DEPLOYED_SC_ADDRESS'", "data": "transfer"}'
echo
echo 'Starting...'

for (( ; ; )); do
response=$(curl -s -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "data": "deploy",
     "scenario": "erc20"
   }' \
  -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple)

echo $response
if ! [[ $response =~ "data" && $response =~ "code" ]] ; then
  echo 'invalid response'
  sleep 10
  continue
fi

if [[ $response =~ "not yet enabled" ]]
then
  echo 'sc deployment not enabled yet.'
  sleep 10
else
  break
fi

done

let COUNT=0

sleep $DEPLOY_SLEEP_DURATION

curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "scenario": "erc20",
     "data": "mint"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple

  sleep $MINT_SLEEP_DURATION

for (( ; ; )); do
  curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "scenario": "erc20",
     "data": "transfer"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple

  sleep $SLEEP_DURATION

  let COUNT++
  if (($COUNT % 5 == 0)); then
    curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "scenario": "erc20",
     "data": "transferToNewAccounts"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple
  fi

done
