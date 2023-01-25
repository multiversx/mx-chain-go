#!/usr/bin/env bash

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

DESTINATION=mixed
if [[ -n $1 ]]; then
  DESTINATION=$1
fi

NUM_TXS=50
VALUE=10
GAS_PRICE=1000000000
GAS_LIMIT=100000
SLEEP_DURATION=40
TOKEN_ISSUE_SLEEP_DURATION=90
MINTING_TXS_SLEEP_DURATION=90

echo 'Starting ESDT scenario cURL script...'
echo 'Sending the token issue transaction and waiting '$TOKEN_ISSUE_SLEEP_DURATION' seconds..'

for (( ; ; )); do
response=$(curl -s -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "data": "mint",
     "scenario": "esdt"
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
  echo 'esdt not enabled yet.'
  sleep 10
else
  break
fi

done

sleep $TOKEN_ISSUE_SLEEP_DURATION

echo 'Sending the minting transactions and waiting '$MINTING_TXS_SLEEP_DURATION' seconds..'
curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "data": "mint",
     "scenario": "esdt"
   }' \
  -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple
sleep $MINTING_TXS_SLEEP_DURATION

echo 'Now the script will start to send regular cURLs'
echo 'Will run a cURL with the following JSON payload each '$SLEEP_DURATION' seconds'
echo '{"value": '$VALUE', "numOfTxs": '$NUM_TXS', "gasPrice": '$GAS_PRICE', "gasLimit": '$GAS_LIMIT', "destination": "'$DESTINATION'", "recallNonce": true, "scenario": "esdt"}'
echo
echo 'Starting...'

for (( ; ; )); do
  curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": true,
     "scenario": "esdt"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple

  sleep $SLEEP_DURATION
done
