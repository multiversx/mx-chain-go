#!/usr/bin/env bash

curl -d '{
  "value": 10,
  "numOfTxs": 2000,
  "numOfShards": 3,
  "gasPrice": 1,
  "gasLimit": 10,
  "crossShard": true,
  "recallNonce": true,
  "scAddress": "000000000000000005009c4a0e096d39eaa6e21c38ffe4262363087622578fcf",
  "data": "transfer"
}' \
  -H "Content-Type: application/json" -X POST http://localhost:7999/transaction/send-multiple
