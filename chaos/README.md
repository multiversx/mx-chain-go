# Chaos testing

The logic within `mx-chain-go/chaos` allows us to test the resilience of the Node by introducing unexpected **failures** within components such as _consensus_, _processing_, etc. The goal of chaos testing is to identify weaknesses and improve the Node's resilience.

## Chaos points

Chaos points are marked in the production code using regular Go comments. For example, in `process/block/shardblock.go`:

```go
func (...) CreateBlock(...) (...) {
	// chaos-testing-point:shardBlock_CreateBlock
    
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

