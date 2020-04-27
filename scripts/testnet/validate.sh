#!/usr/bin/env bash

# Call a REST endpoint of the TxGen, requesting it to query the Proxy for all
# known account balances and nonces, in order to compare their expected state
# (stored internally in TxGen) with the actual ones from the testnet.

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

if [ $USE_TXGEN -eq 0 ]; then
  echo "No validation to perform: TxGen is disabled."
else
  if [ $TXGEN_ERC20_MODE -eq 0 ]; then
    echo "Validating native accounts (non-ERC20 mode)"
    RESULT=$(curl -s "http://127.0.0.1:$PORT_TXGEN/validate/basic")
  else
    echo "Validating ERC20 accounts"
    DEPLOYED_SC_ADDRESS=$(cat $TESTNETDIR/txgen/deployedSCAddress.txt)
    RESULT=$(curl -s "http://127.0.0.1:$PORT_TXGEN/validate/sc/$DEPLOYED_SC_ADDRESS")
  fi

  echo "$RESULT" | jq .

  validation_errors=$(echo "$RESULT" | jq . | grep ': {' | wc -l)
  echo "Validation result: $validation_errors errors"
fi
