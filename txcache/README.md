## Mempool

### Glossary

1. **selection session:** an ephemeral session during which the mempool selects transactions for a proposer. A session starts when a proposer asks the mempool for transactions and ends when the mempool returns the transactions. The most important part of a session is the _selection loop_.
2. **transaction PPU:** the price per unit of computation, for a transaction. It's computed as `initiallyPaidFee / gasLimit`.
3. **initially paid transaction fee:** the fee for processing a transaction, as known before its actual processing. That is, without knowing the _refund_ component.

### Configuration

1. **SelectTransactions::gasRequested:** `10_000_000_000`, the maximum total gas limit of the transactions to be returned to a proposer (one _selection session_). This value is provided by the Protocol.
2. **SelectTransactions::maxNum:** `50_000`, the maximum number of transactions to be returned to a proposer (one _selection session_). This value is provided by the Protocol.

### Transactions selection

### Paragraph 1

When a proposer asks the mempool for transactions, it provides the following parameters:

 - `gasRequested`: the maximum total gas limit of the transactions to be returned
 - `maxNum`: the maximum number of transactions to be returned

### Paragraph 2

The PPU (price per gas unit) of a transaction, is computed (once it enters the mempool) as follows:

```
ppu = initiallyPaidFee / gasLimit
```

In the formula above, 

```
initiallyPaidFee =
    dataCost * gasPrice +
    executionCost * gasPrice * network.gasPriceModifier

dataCost = network.minGasLimit + len(data) * network.gasPerDataByte

executionCost = gasLimit - dataCost
```

#### Examples

(a) A simple native transfer with `gasLimit = 50_000` and `gasPrice = 1_000_000_000`:

```
initiallyPaidFee = 50_000_000_000 atoms
ppu = 1_000_000_000 atoms
```

(b) A simple native transfer with `gasLimit = 50_000` and `gasPrice = 1_500_000_000`:

```
initiallyPaidFee = gasLimit * gasPrice = 75_000_000_000 atoms
ppu = 75_000_000_000 / 50_000 = 1_500_000_000 atoms
```

(c) A simple native transfer with a data payload of 7 bytes, with `gasLimit = 50_000 + 7 * 1500` and `gasPrice = 1_000_000_000`:

```
initiallyPaidFee = 60_500_000_000_000 atoms
ppu = 60_500_000_000_000 / 60_500 = 1_000_000_000 atoms
```

That is, for simple native transfers (whether they hold a data payload or not), the PPU is equal to the gas price.

(d) ...

### Paragraph 4

Transaction **A** is considered more desirable (for the Network) than transaction **B** if **it has a higher PPU**.

If two transactions have the same PPU, they are ordered using an arbitrary, but deterministic rule: the transaction with the higher [fvn32(transactionHash)](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function) "wins" the comparison.

Pseudo-code:

```
func isTransactionMoreDesirableToNetwork(A, B):
    if A.ppu > B.ppu:
        return true
    if A.ppu < B.ppu:
        return false
    return fvn32(A.hash) > fvn32(B.hash)
```

### Paragraph 3

The mempool selects transactions as follows:
 - before starting the selection loop, get a snapshot of the senders, in an arbitrary order.
 - in the selection loop, do as many _passes_ as needed to satisfy `gasRequested` (see **Paragraph 1**).
 - within a _pass_, ...
 - if `gasRequested` is satisfied, stop the _selection loop_ early.
 - if `maxNum` is satisfied, stop the _selection loop_ early.

### Paragraph 4

Within a _selection pass_, a batch of transactions from a sender is selected as follows:
 - ..., attempt to **detect an initial nonces gap** (if enough information is available, that is, if the current account nonce is known - see section **Account nonce notifications**).
 - if a nonces gap is detected, ... Subsequent passes of the selection loop (within the same selection session) will skip this sender. The sender will be re-considered in a future selection session.

#### Initial gaps and middle gaps

### Account nonce notifications

### Transactions addition

### Transactions removal

### Transactions eviction

### Monitoring and diagnostics

