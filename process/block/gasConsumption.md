## Gas consumption implementation

### 1. High level overview

A new gas computation component is needed throughout the block creation mechanism, during transactions selection.

The main purpose of this component will be to help filling the block efficiently, in respect with the gas limits defined in config.

First, incoming miniblocks referenced by metachain will be added. Each miniblock will be verified against `MaxGasLimitPerMiniBlock` then the new component will check if the current miniblock can be added in the initial limit of 50% of `MaxGasLimitPerBlock`. Once the limits are reached, the component will not accept other miniblocks.

Second, the outgoing list of transactions will be checked agains `MaxGasLimitPerTx`, then `MaxGasLimitPerMiniBlock` and the other 50% of `MaxGasLimitPerBlock`. If there is not enough "space" for all transactions, the component will return the last transaction index added.

Throughout the selection process, up until the next explicit reset called, gasComputation component will store the consumed gas.

After both incoming miniblocks and outgoing transactions are added, it should offer the option to provide the free gas left in the block and also allow appending either more miniblocks, either more transactions, in order to fill the block as much as possible.

Another feature that must be satisfied is the option to decrease the limits when needed. Decreasing limits and resetting them should be done separately.

### 2. Proposed interface

```go
type GasComputation interface {
    CheckIncomingMiniBlock(
        miniBlock data.MiniBlockHeaderHandler,
        transactions []data.TransactionHandler,
    ) error
    CheckOutgoingTransactions(transactions []data.TransactionHandler) uint32
    TotalGasConsumed() uint64
    IsBlockFull() bool
    DecreaseMiniBlockLimit()
    ResetMiniBlockLimit()
    DecreaseBlockLimit()
    ResetBlockLimit()
    Reset()
}
```

### 3. Proposed implementation details

- `CheckIncomingMiniBlock(miniBlock data.MiniBlockHeaderHandler, transactions []data.TransactionHandler) error`
  - receives an incoming miniBlock and the transactions included in the miniBlock
  - iterates through the provided transactions and:
    - checks tx limit
    - computes the total gas in the miniBlock
    - checks the miniBlock limit
    - checks if the miniBlock can be included without reaching the block limit
  - returns error if any limit is reached or nil if the miniBlock can be included in the current block
- `CheckOutgoingTransactions(transactions []data.TransactionHandler) uint32`
  - receives a list of outgoing transactions
  - iterates through the received transactions and:
    - checks tx limit
    - computes the total gas of the transactions
    - checks the miniBlock limit
    - checks if the transactions can all be included in the block
  - returns the index of the last transaction that can be included without reaching limits
- `TotalGasConsumed() uint64`
  - returns the total gas consumed so far
- `IsBlockFull() bool`
  - returns true if the current block does not accept any transaction
- `DecreaseMiniBlockLimit()`
  - decreases the miniBlock limit by a pre-configured percent
  - once modified, it will stay decreased until manual reset
- `ResetMiniBlockLimit()`
  - resets the miniBlock limit to the initial value
- `DecreaseBlockLimit()`
  - decreases the block limit by a pre-configured percent
  - once modified, it will stay decreased until manual reset
- `ResetBlockLimit()`
  - resets the block limit to the initial value
- `Reset()`
  - resets the gas consumed for the current block

The total block limit will be equally distributed among incoming miniBlocks and outgoing transactions.

Considering the flow where the incoming miniBlocks are handled first, the gasComputation component will only accept miniBlocks until 50% of the block limit.
Then, when outgoing transactions are handled, they will also be added up until the other 50% of the block limit.
If the outgoing transactions do not fill the entire 50%, the component will allow further calls to `CheckIncomingMiniBlock` to add more miniBlocks over the 50% initially considered as limit.

Similar treatment will be considered on the second case, when outgoing transactions are handled first, then incoming miniBlocks.
After finishing the incoming miniBlocks, if there is more space left, more outgoing transactions can be added via `CheckOutgoingTransactions`.

### 4. Structure

```go
type gasConsumption struct {
    sync.RWMutex
    economicsFee           process.FeeHandler
    totalGasConsumed       map[gasType]uint64
    gasConsumedByMiniBlock map[string]uint64
    miniBlockLimitFactor   uint64
    blockLimitFactor       uint64
}
```
where gasType is defined as follows:
```go
type gasType string

const (
	incomingTx gasType = "incoming"
	outgoingTx gasType = "outgoing"
)
```
- extends RWMutex, in order to protect the internal maps and limits
- `economicsFee` will provide access to the configured limits
- `totalGasConsumed` holds the gas consumed by incoming/outgoing transactions
- `gasConsumedByMiniBlock` holds the gas consumed by each miniBlock (for further requests, might not be needed)
- `miniBlockLimitFactor` holds the factor that will be multiplied with the limits miniBlock limit received from economicsFee instance (initial 100%, can be decreased via `DecreaseMiniBlockLimit`)
- `blockLimitFactor` holds the factor that will be multiplied with the limits block limit received from economicsFee instance (initial 100%, can be decreased via `DecreaseBlockLimit`)
