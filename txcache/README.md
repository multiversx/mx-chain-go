## Mempool

### Glossary

1. **selection session:** an ephemeral session during which the mempool selects transactions for a proposer. A session starts when a proposer asks the mempool for transactions and ends when the mempool returns the transactions. The most important part of a session is the _selection loop_.
2. **selection pass:** a single iteration of the _selection loop_. In an iteration, the algorithm goes through all the senders (appropriately sorted) and selects a batch of transactions from each sender. A _pass_ can stop early (see **Paragraph 3**).
3. **sender score:** a score assigned to a sender based on her's behavior. The score is used to determine the order in which senders are considered within a _selection pass_, plus the size and capacity of a _sender's transactions batch_. The score is a number in `[0, maxSenderScore]`.

### Configuration

1. **maxSenderScore:** `100`, the maximum score a sender can have. The minimum score is `0`.
3. **gasRequested:** `10_000_000_000`, the maximum total gas limit of the transactions to be returned to a proposer (one _selection session_).
4. **baseNumPerSenderBatch:**: `100`, defines the maximum number of transactions to be selected from the transactions pool, for a sender with the maximum possible score, in a _single pass_. Senders with lower scores will have fewer transactions selected in a single pass.
5. **baseGasPerSenderBatch:**: `120_000_000`, defines the maximum gas for transactions to be selected from the transactions pool, for a sender with the maximum possible score, in a single pass. Senders with lower scores will have less gas selected in a single pass.

### Transactions selection

### Paragraph 1

When a proposer asks the mempool for transactions, it provides the following parameters:

 - `gasRequested`: the maximum total gas limit of the transactions to be returned
 - `baseNumPerSenderBatch`: a base value for the number of transactions to be returned per sender, per selection _pass_. This value is used to compute the actual number of transactions to be returned per sender, per selection _pass_, based on the sender's score (see **Paragraph 2**).
 - `baseGasPerSenderBatch`: a base value for the total gas limit of the transactions to be returned per sender, per selection _pass_. This value is used to compute the actual total gas limit of the transactions to be returned per sender, per selection _pass_, based on the sender's score (see **Paragraph 2**). Due to how the selection is performed, the theoretical maximum gas might be exceeded (a bit), as follows: `theoretical maximum = (baseGasPerSenderBatch - 1) + max(baseGasPerSenderBatch, max gas limit of a transaction)`. Think of a sender with maximum score, having two transactions, one with `gasLimit = baseGasPerSenderBatch - 1`, and the other with `gasLimit = max gas limit of a transaction`.

### Paragraph 2

How is the size of a sender batch computed?

1. If the score of the sender is **zero**, then the size of the sender batch is **1**, and the total gas limit of the sender batch is **1**.
2. If the score of the sender is **non-zero**, then the size of the sender batch is computed as follows:
   - `scoreDivision = score / maxSenderScore`
   - `numPerBatch = baseNumPerSenderBatch * scoreDivision`
   - `gasPerBatch = baseGasPerSenderBatch * scoreDivision`

Examples:
 - for `score == 100`, we have `numPerBatch == 100` and `gasPerBatch == 120000000`
 - for `score == 74`, we have `numPerBatch == 74` and `gasPerBatch == 88800000`
 - for `score == 1`, we have `numPerBatch == 1` and `gasPerBatch == 1200000`
 - for `score == 0`, we have `numPerBatch == 1` and `gasPerBatch == 1`

### Paragraph 3

The mempool selects transactions as follows:
 - before starting the selection loop, get a snapshot of the senders (sorted by score, descending)
 - in the selection loop, do as many _passes_ as needed to satisfy `gasRequested` (see **Paragraph 1**).
 - within a _pass_, go through all the senders (appropriately sorted) and select a batch of transactions from each sender. The size of the batch is computed as described in **Paragraph 2**.
 - if `gasRequested` is satisfied, stop the _pass_ early.

### Paragraph 4

Within a _selection pass_, a batch of transactions from a sender is selected as follows:
 - if it's the first pass, then reset the internal state used for copy operations (in the scope of a sender). Furthermore, attempt to **detect an initial nonces gap** (if enough information is available, that is, if the current account nonce is known - see section **Account nonce notifications**).
 - if a nonces gap is detected, return an empty batch. Subsequent passes of the selection loop (within the same selection session) will skip this sender. The sender will be re-considered in a future selection session.
 - go through the list of transactions of the sender (sorted by nonce, ascending) and select the first `numPerBatch` transactions that fit within `gasPerBatch`.
 - in following passes (within the same selection session), the batch selection algorithm will continue from the last selected transaction of the sender (think of it as a cursor).

### Score computation

The score of a sender it's computed based on her transactions (as found in the mempool) and the account nonce (as learned through the _account nonce notifications_).

The score is strongly correlated with the average price paid by the sender per unit of computation - we'll call this **avgPpu**, as a property of the sender.

Additionally, we define two global properties: `worstPpu` and `excellentPpu`. A sender with an `avgPpu` of `excellentPpu + 1` gets the maximum score, while a sender with an `avgPpu` of `worstPpu` gets the minimum score.

`worstPpu` is computed as the average price per unit of the "worst" possible transaction - minimum gas price, maximum gas limit, and minimum data size (thus abusing the Protocol gas price subvention):

```
worstPpu = (50000 * 1_000_000_000 + (600_000_000 - 50000) * (1_000_000_000 / 100)) / 600_000_000
         = 10082500
```

`excellentPpu` is set to `minGasPrice` times a _chosen_ factor:

```
excellentPpu = 1_000_000_000 * 5 = 5_000_000_000
```

Examples:
 - ...

#### Spotless sequence of transactions

### Account nonce notifications

### Transactions addition

### Transactions removal

### Transactions eviction

### Monitoring and diagnostics

