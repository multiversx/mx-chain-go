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

Network parameters (as of November of 2024):
    
```
gasPriceModifier = 0.01
minGasLimit = 50_000
gasPerDataByte = 1_500
```

#### Examples

**(a)** A simple native transfer with `gasLimit = 50_000` and `gasPrice = 1_000_000_000`:

```
initiallyPaidFee = 50_000_000_000 atoms
ppu = 1_000_000_000 atoms
```

**(b)** A simple native transfer with `gasLimit = 50_000` and `gasPrice = 1_500_000_000`:

```
initiallyPaidFee = gasLimit * gasPrice = 75_000_000_000 atoms
ppu = 75_000_000_000 / 50_000 = 1_500_000_000 atoms
```

**(c)** A simple native transfer with a data payload of 7 bytes, with `gasLimit = 50_000 + 7 * 1500` and `gasPrice = 1_000_000_000`:

```
initiallyPaidFee = 60_500_000_000_000 atoms
ppu = 60_500_000_000_000 / 60_500 = 1_000_000_000 atoms
```

That is, for simple native transfers (whether they hold a data payload or not), the PPU is equal to the gas price.

**(d)** A contract call with `gasLimit = 75_000_000` and `gasPrice = 1_000_000_000`, with a data payload of `42` bytes:

```
initiallyPaidFee = 861_870_000_000_000 atoms
ppu = 11_491_600 atoms
```

**(e)** Similar to **(d)**, but with `gasPrice = 2_000_000_000`:

```
initiallyPaidFee = 1_723_740_000_000_000 atoms
ppu = 22_983_200 atoms
```

That is, for contract calls, the PPU is not equal to the gas price, but much lower, due to the contract call _cost subsidy_. **A higher gas price will result in a higher PPU.**

### Paragraph 3

Transaction **A** is considered **more valuable (for the Network)** than transaction **B** if **it has a higher PPU**.

If two transactions have the same PPU, they are ordered using an arbitrary, but deterministic rule: the transaction with the "lower" transaction hash "wins" the comparison.

Pseudo-code:

```
func isTransactionMoreValuableForNetwork(A, B):
    if A.ppu > B.ppu:
        return true
    if A.ppu < B.ppu:
        return false
    return A.hash < B.hash
```

### Paragraph 4

The mempool selects transactions as follows (pseudo-code):

```
func selectTransactions(gasRequested, maxNum):
    // Setup phase
    senders := list of all current senders in the mempool, in an arbitrary order
    bunchesOfTransactions := sourced from senders; middle-nonces-gap-free, duplicates-free, nicely sorted by nonce

    // Holds selected transactions
    selectedTransactions := empty

    // Holds not-yet-selected transactions, ordered by PPU
    competitionHeap := empty
    
    for each bunch in bunchesOfTransactions:
        competitionHeap.push(next available transaction from bunch)
    
    // Selection loop
    while competitionHeap is not empty:
        mostValuableTransaction := competitionHeap.pop()

        // Check if adding the next transaction exceeds limits
        if selectedTransactions.totalGasLimit + mostValuableTransaction.gasLimit > gasRequested:
            break
        if selectedTransactions.length + 1 > maxNum:
            break
        
        selectedTransactions.append(mostValuableTransaction)
        
        nextTransaction := next available transaction from the bunch of mostValuableTransaction
        if nextTransaction exists:
            competitionHeap.push(nextTransaction)

    return selectedTransactions
```

Thus, the mempool selects transactions using an efficient and value-driven algorithm that ensures the most valuable transactions (in terms of PPU) are prioritized while maintaining correct nonce sequencing per sender. The selection process is as follows:

**Setup phase:**

   - **Snapshot of senders:**
     - Before starting the selection loop, obtain a snapshot of all current senders in the mempool in an arbitrary order.

   - **Organize transactions into bunches:**
     - For each sender, collect all their pending transactions and organize them into a "bunch."
     - Each bunch is:
       - **Middle-nonces-gap-free:** There are no missing nonces between transactions.
       - **Duplicates-free:** No duplicate transactions are included.
       - **Sorted by nonce:** Transactions are ordered in ascending order based on their nonce values.

   - **Prepare the heap:**
     - Extract the first transaction (lowest nonce) from each sender's bunch.
     - Place these transactions onto a max heap, which is ordered based on the transaction's PPU.

**Selection loop:**

   - **Iterative selection:**
     - Continue the loop until either the total gas of selected transactions meets or exceeds `gasRequested`, or the number of selected transactions reaches `maxNum`.
     - In each iteration:
       - **Select the most valuable transaction:**
         - Pop the transaction with the highest PPU from the heap.
         - Append this transaction to the list of `selectedTransactions`.
       - **Update the sender's bunch:**
         - If the sender of the selected transaction has more transactions in their bunch:
           - Take the next transaction (next higher nonce) from the bunch.
           - Push this transaction onto the heap to compete in subsequent iterations.
     - This process ensures that at each step, the most valuable transaction across all senders is selected while maintaining proper nonce order for each sender.

   - **Early termination:**
     - The selection loop can terminate early if either of the following conditions is satisfied before all transactions are processed:
       - The accumulated gas of selected transactions meets or exceeds `gasRequested`.
       - The number of selected transactions reaches `maxNum`.

**Additional notes:**
 - Within the selection loop, the current nonce of the sender is queryied from the blockchain, if necessary.
 - If an initial nonce gap is detected, the sender is excluded from the selection process.
 - Transactions with nonces lower than the current nonce of the sender are skipped (not included in the selection).

### Paragraph 5

On the node's side, the selected transactions are shuffled using a deterministic algorithm. This shuffling ensures that the transaction order remains unpredictable to the proposer, effectively preventing _front-running attacks_. Therefore, being selected first by the mempool does not guarantee that a transaction will be included first in the block. Additionally, selection by the mempool does not ensure inclusion in the very next block, as the proposer has the final authority on which transactions to include, based on **the remaining space available** in the block.
