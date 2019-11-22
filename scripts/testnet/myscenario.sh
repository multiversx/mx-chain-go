#!/usr/bin/env bash

curl -d '{
  "value": 10,
  "numOfTxs": 2000,
  "numOfShards": 3,
  "gasPrice": 1,
  "gasLimit": 10,
  "crossShard": true,
  "recallNonce": true,
  "scAddress": "00000000000000000500b753a538d157677a8a18e97afa92460c5efaec3ded99",
  "data": "transfer"
}' \
  -H "Content-Type: application/json" -X POST http://localhost:7999/transaction/send-multiple
