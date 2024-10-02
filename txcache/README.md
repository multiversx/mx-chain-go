## Mempool

### Glossary

1. **maxSenderScore:** 100 (constant)
2. **selection pass:** a single iteration of the _selection loop_. In a single iteration, the algorithm goes through all the senders (appropriately sorted) and selects a batch of transactions from each sender. A _pass_ can stop early (see **Paragraph 3**).

### Transactions selection

### Paragraph 1

When a proposer asks the mempool for transactions, it provides the following parameters:

 - `numRequested`: the maximum number of transactions to be returned
 - `gasRequested`: the maximum total gas limit of the transactions to be returned
 - `baseNumPerSenderBatch`: a base value for the number of transactions to be returned per sender, per selection _pass_. This value is used to compute the actual number of transactions to be returned per sender, per selection _pass_, based on the sender's score (see **Paragraph 2**).
 - `baseGasPerSenderBatch`: a base value for the total gas limit of the transactions to be returned per sender, per selection _pass_. This value is used to compute the actual total gas limit of the transactions to be returned per sender, per selection _pass_, based on the sender's score (see **Paragraph 2**).

### Paragraph 2

How is the size of a sender batch computed?

1. If the score of the sender is **zero**, then the size of the sender batch is **1**, and the total gas limit of the sender batch is **1**.
2. If the score of the sender is **non-zero**, then the size of the sender batch is computed as follows:
   - `scoreDivision = score / maxSenderScore`
   - `numPerBatch = baseNumPerSenderBatch * scoreDivision`
   - `gasPerBatch = baseGasPerSenderBatch * scoreDivision`

### Paragraph 3

The mempool selects transactions as follows:
 - before starting the selection loop, get a snapshot of the senders (sorted by score, descending)
 - in the selection loop, do as many passes as needed to satisfy `numRequested` and `gasRequested` (see **Paragraph 1**).

### Transactions addition

### Transactions removal

### Monitoring and diagnostics

