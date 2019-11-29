export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

if [ $PRIVATE_REPOS -eq 0 ]; then
  echo "No scenarios can be run if private repositories are disabled."
else
  sendTransactions() {
    curl -d '{
      "value": 10,
      "numOfTxs": 200,
      "numOfShards": '$SHARDCOUNT',
      "gasPrice": 1,
      "gasLimit": 200000,
      "crossShard": true,
      "recallNonce": true
    }' \
      -H "Content-Type: application/json" -X POST http://localhost:$PORT_TXGEN/transaction/send-multiple
    echo ""
  }

  while true
  do
    sendTransactions
    echo ""
    echo "Transaction batch emitted"
    sleep 60

    $ELRONDTESTNETSCRIPTSDIR/validate.sh
  done
fi
