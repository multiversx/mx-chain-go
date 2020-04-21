# Transactions cache

### Pool structure

The transactions pool contains more caches, as follows:

 - `1` cache structure to hold all transactions where `source == me`
 - `N - 1` cache structures, one for each possible source shard where `destination == me`, but `source != me`

where `N` is the number of shards.

### Cache structure

### Cache capacity

The first cache (the union) is `N` times larger than the others. Given `N = 5` and the configuration:

```
[TxDataPool]
    Size = 900000
    SizeInBytes = 524288000
    ...
```

it follows that the first cache has a capacity of `500000` transactions (or `~277 MB`) while the rest have a capacity of `100000` transactions (or `~56 MB`) each.

When capacity is close to be reached, **eviction** is triggered.

### Insertion in cache

 1. The eviction condition is tested before the actual addition.
 1. If eviction is necessary, it is executed synchronously.
 1. The incoming transaction is added in the cache if missing.

### Selection of transactions

The selection is called by the processing components. Typically, the selection buffer has a size of `numRequested = 30000` transactions and the base count, sender-scoped, is `batchSizePerSender = 10`.

 1. Once started, the selection stops when the buffer is full or when there are no more transactions to select
 1. Selection is performed in a loop, and its iterations are called **selection passes**
 1. At each selection pass, the senders are taken in **an approximate order** from the highest **score** to the lowest, and given the opportunity to contribute transactions. The bulk of transactions is called **contribution**. One *contribution* is scoped to one *selection pass*. Since there are *many selection passes*, a given sender could have *many contributions* during a single *selection*.
 1. One sender can contribute at most `batchSizePerSender * (score + 1)` in one selection pass. For example, a sender with `score == 100` could have a *contribution* of `10 * 101 = 1010` transactions is one selection pass.

#### Contributing transactions

In each **selection pass*, a sender may contribute transactions. The contribution itself depends on several factors.

 1. Transactions are contributed in the **nonce** order
 1. In the first *selection pass* the sender is checked for an **initial nonce gap**
 1. If there is an *initial nonce gap*, the sender is blocked any contribution in the current selection, and the sender is tagged with `failedSelection`. All subsequent *selection passes* in the current selection will ignore the sender
 1. If the sender is tagged with `failedSelection` several (`5`) times in a row, it enters the **grace period**
 1. The *grace period* of the sender expires on its `7th` `failedSelection` in a row
 1. During the *grace period* (`5` to `7` failed selections), the sender is permitted a contribution of *one transaction* per selection, which is called the **grace transaction**
 1. Once the *initial nonce gap* is resolved, the `failedSelection` tag is removed
 1. If, when contributing transactions, a *nonce gap* (called a *middle nonce gap*) is encountered, the sender is blocked any contribution in the current selection

 ### Score of senders

 ### Eviction