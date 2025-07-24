## Gas consumption implementation

### 1. High level overview

A new gas computation component is needed throughout the block creation mechanism, during transactions selection.

The main purpose of this component will be to help filling the block efficiently, in respect with the gas limits defined in config.

First, incoming mini blocks referenced by metachain will be added. Each mini block will be verified against `MaxGasLimitPerMiniBlock` then the new component will check if each mini block can be added in the initial limit of 50% of `MaxGasLimitPerBlock`. Once the limits are reached, the component will store the remaining mini blocks as pending, until outgoing transactions are handled.

Second, the outgoing list of transactions will be checked against `MaxGasLimitPerTx`, then against the remaining 50% of `MaxGasLimitPerBlock`. If there is not enough "space" for all transactions, the component will first verify if there are any pending mini blocks from the above step. If none, it will continue to add transactions. Otherwise, it will return the last index added.

If the outgoing transactions didn't fill the entire 50% space allocated for them, this component will automatically try to add more pending mini blocks if any, until the block is filled.

Throughout the selection process, up until the next explicit reset called, gasComputation component will store the consumed gas.

Another feature that must be satisfied is the option to decrease the limits when needed. Decreasing limits and resetting them should be done separately.

### 2. Proposed interface

```go
type GasComputation interface {
    CheckIncomingMiniBlocks(
        miniBlocks   []data.MiniBlockHeaderHandler,
        transactions []data.TransactionHandler,
    ) (uint32, uint32)
    CheckOutgoingTransactions(transactions []data.TransactionHandler) uint32
    TotalGasConsumed() uint64
    GetLastMiniBlockIndexIncluded() (bool, uint32)
    GetLasTransactionIndexIncluded() (bool, uint32)
    DecreaseMiniBlockLimit()
    ResetMiniBlockLimit()
    DecreaseBlockLimit()
    ResetBlockLimit()
    Reset()
}
```

### 3. Proposed implementation details

- `CheckIncomingMiniBlocks(miniBlocks []data.MiniBlockHeaderHandler, transactions []data.TransactionHandler) error`
  - receives a list of incoming mini blocks and the transactions included in them
  - iterates through the provided transactions and:
    - checks tx limit
    - computes the total gas in the mini block
    - checks the mini block limit
    - checks if the mini block can be included without reaching the block limit
  - if `MaxGasLimitPerMiniBlock` is reached, it stores the remaining mini blocks as pending
  - if outgoing transactions are already handled and there is still some space left, it will continue to add the pending mini blocks
  - returns the number of mini blocks included and the number of remaining pending
- `CheckOutgoingTransactions(transactions []data.TransactionHandler) uint32`
  - receives a list of outgoing transactions
  - iterates through the received transactions and:
    - checks tx limit
    - computes the total gas of the transactions
    - checks the miniBlock limit
    - checks if the transactions can all be included in the block
  - if 50% of the `MaxGasLimitPerBlock` is reached, it will first check if there are any pending mini blocks and incoming mini blocks were previously. If none, it will continue to add transactions until the block is full.
  - if incoming mini blocks were not handled yet, it will store the remaining transactions as pending transactions
  - returns the index of the last transaction that can be included without reaching limits
- `TotalGasConsumed() uint64`
  - returns the total gas consumed so far
- `GetLastMiniBlockIndexIncluded() (bool, uint32)`
  - returns the last mini block index that was included and a true boolean if the selection is over. If there are still any pending mini blocks, it will return false.
- `GetLasTransactionIndexIncluded() (bool, uint32)`
  - returns the last transaction index that was included and a true boolean if the selection is over. If there are still any pending transactions, it will return false.
- `DecreaseMiniBlockLimit()`
  - decreases the mini block limit by a pre-configured percent
  - once modified, it will stay decreased until manual reset
- `ResetMiniBlockLimit()`
  - resets the mini block limit to the initial value
- `DecreaseBlockLimit()`
  - decreases the block limit by a pre-configured percent
  - once modified, it will stay decreased until manual reset
- `ResetBlockLimit()`
  - resets the block limit to the initial value
- `Reset()`
  - resets the gas consumed for the current block

The total block limit will be equally distributed among incoming mini blocks and outgoing transactions.

Considering the flow where the incoming mini blocks are handled first, the gasComputation component will only accept mini blocks until 50% of the block limit.
Then, when outgoing transactions are handled, they will also be added up until the other 50% of the block limit.
If the outgoing transactions do not fill the entire 50%, the component will allow further calls to `CheckIncomingMiniBlock` to add more mini blocks over the 50% initially considered as limit.

Similar treatment will be considered on the second case, when outgoing transactions are handled first, then incoming mini blocks.
After finishing the incoming mini blocks, if there is more space left, more outgoing transactions can be added via `CheckOutgoingTransactions`.

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
- `gasConsumedByMiniBlock` holds the gas consumed by each mini block (for further requests, might not be needed)
- `miniBlockLimitFactor` holds the factor that will be multiplied with the limits mini block limit received from economicsFee instance (initial 100%, can be decreased via `DecreaseMiniBlockLimit`)
- `blockLimitFactor` holds the factor that will be multiplied with the limits block limit received from economicsFee instance (initial 100%, can be decreased via `DecreaseBlockLimit`)
