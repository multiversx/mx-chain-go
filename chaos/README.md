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
python3 ./chaos/insert_chaos.py
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

### `creatingBlockError`

### `processingBlockError`

### `consensusCorruptSignature`

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

### `consensusV2CorruptLeaderSignature`

### `consensusV2DelayLeaderSignature`

### `consensusV2SkipSendingBlock`
