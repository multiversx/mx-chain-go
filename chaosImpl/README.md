# Chaos testing

The logic within `mx-chain-go/chaos` allows us to test the resilience of the Node by introducing unexpected **failures** within components such as _consensus_, _processing_, etc. The goal of chaos testing is to identify weaknesses and improve the Node's resilience.

## Chaos points

Chaos points are marked in the production code using regular Go comments. For example, in `process/block/shardblock.go`:

```go
func (...) CreateBlock(...) (...) {
	// chaos:shardBlockCreateBlock
    
    ...
```

During CI pipelines or testing procedures, **failures** can be introduced at these points by running the following Python script, before building the Node:

```
python3 ./chaosImpl/insert_chaos.py --config-file=cmd/node/config/chaos.json
```

The logic of the _failures_ themselves is defined in the `chaos` package. See `chaos/controller.go`.

Developers are responsible for adding new chaos points, when necessary, by extending the `chaos` package and adding the corresponding markers (comments) in the production code.

## Configuration

Failures are configured in `cmd/node/config/chaos.json`. Let's take the example of the failure `creatingBlockError`:

```json
{
    "failures": [
        {
            "name": "creatingBlockError",
            "enabled": true,
            "triggers": ["shard == 0 && round == 20"]
        },
        ...
    ]
```

## Failures

### `consensusV1ReturnErrorInCheckSignaturesValidity`

In subround `END ROUND`, in `checkSignaturesValidity`, return an early error. Note that this failure is available both for leaders and for validators. Adjust the failure triggers accordingly. For example:

```json
{
    "name": "consensusV1ReturnErrorInCheckSignaturesValidity",
    "enabled": true,
    "triggers": [
        "!iAmLeader && round % 11 == 0 && nodeIndex < consensusSize / 3",
        "iAmLeader && round % 13 == 0"
    ]
}
```

### `consensusV1DelayBroadcastingFinalBlockAsLeader`

This failure is internally known as _testnet soft forks_. In subround `END ROUND`, as a leader, delay broadcasting the final block. For example:

```json
{
    "name": "consensusV1DelayBroadcastingFinalBlockAsLeader",
    "enabled": true,
    "triggers": [
        "round % 20 == 0"
    ]
}
```


## Alter chaos configuration using transactions

```
export WALLET_0="~/multiversx-sdk/testwallets/latest/users/bob.pem"
export WALLET_1="~/multiversx-sdk/testwallets/latest/users/alice.pem"
export WALLET_2="~/multiversx-sdk/testwallets/latest/users/carol.pem"
export WALLET_METACHAIN="~/multiversx-sdk/testwallets/latest/users/judy.pem"

export ADDRESS_0="erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx"
export ADDRESS_1="erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
export ADDRESS_2="erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8"
export ADDRESS_METACHAIN="erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"

export PROXY="http://localhost:7950"
```

Select a chaos profile (per shard):

```
cat << EOF > transaction-data.json
{
    "action": "selectProfile",
    "profile": "default"
}
EOF

mxpy tx new --data-file=transaction-data.json --pem=$WALLET_0 --gas-limit 5000000 --recall-nonce --receiver=$ADDRESS_0 --proxy=$PROXY --send
mxpy tx new --data-file=transaction-data.json --pem=$WALLET_1 --gas-limit 5000000 --recall-nonce --receiver=$ADDRESS_1 --proxy=$PROXY --send
mxpy tx new --data-file=transaction-data.json --pem=$WALLET_2 --gas-limit 5000000 --recall-nonce --receiver=$ADDRESS_2 --proxy=$PROXY --send
mxpy tx new --data-file=transaction-data.json --pem=$WALLET_METACHAIN --gas-limit 5000000 --recall-nonce --receiver=$ADDRESS_METACHAIN --proxy=$PROXY --send
```
