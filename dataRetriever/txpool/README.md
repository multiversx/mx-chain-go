# Transactions Pool

### Pool structure

The transactions pool contains more caches, as follows:

 - `1` cache structure to hold all transactions where `source == me`. This structure is of type `txcache.TxCache` and we call it **TxCache**.
 - `N - 1` cache structures, one for each possible source shard where `destination == me`, but `source != me`. These structures are of type `txcache.CrossTxCache` and we call them **CrossTxCaches**.

Above, `N` is the number of shards.

### Cache functions

 1. **insertion (both caches):** this function is primarly required by the **incerceptor processor**, which adds transactions to the pool, upon validating them.
 
 1. **selection (`TxCache`):** this function retrieves a number (e.g. 30000) of transactions from the **TxCache** so that they are further fed to the transaction processing components. This function is required **on "my turn"** only.
 
 1. **removal (both caches):** this function is required by components such as the **transactions cleaner** (which cleans old, stale transactions from time to time), the **transactions processor** (which removes bad transactions at processing time), the **base processor** (which removes transactions included in commited blocks) etc.

 1. **eviction (both caches):** when capacity is closed to be reached, the caches heuristically discard bulks of transactions. The high-load conditions are checked at the time of insertion.

 1. **immunization (`CrossTxCache`):** this function is invoked by the **block tracker** when a miniblock containing cross-shard transactions is received. The transactions listed in the miniblock are must-have, non-evictable transactions, since they will be soon searched for in the pool, for processing.

### Cache structure

#### TxCache

The `txcache.TxCache` is backed by two maps, one to allow quick arbitrary access to transactions by their hash (`txByHashMap`) and the other one to allow quick **selection** of transactions eligible for processing (`txListBySenderMap`). The second map holds one sorted list of transactions for each sender - transactions are **sorted by their nonce**, and these lists are also kept approximately sorted with respect to a so-called **sender score** (see below). 

These two maps are kept in mutual **consistency** (e.g. a transaction is present in both maps or in no map). Slight inconsistencties might occur - they have been identified, documented in the code, and they do not cause severe issues. Insertions and deletions are not performed within a critical section (to avoid locking in this highly concurrent environment) and are not atomic. Mutation operations (removals and additions) are sequenced so that consistency is eventually reached.

#### CrossTxCache

A `txcache.CrossTxCache` is backed by a special type of cache called `ImmunityCache`. This one allows one to immunize a set of items against eviction.

#### Concurrency

All backing map-like structures are **concurrent-safe**, and are also `sharded` (or `chunked`) so that locking is local (as opposed to global) and precise. Chunk affinity of items (transactions, senders) is decided using the [Fowler–Noll–Vo hash function](https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function). 

#### Snapshoting in `TxCache`

For both **transaction selection** and **eviction**, the `txcache.TxCache` uses a _data snapshotting_ technique. More specifically:

 1. at selection time, a snapshot of all the senders in the cache (sorted in _descending_ score order) is created. This snapshot is then iterated over and eligible transactions are thus selected from theoretically soon-to-be or possibly already stale data. In practice, this is not an issue, since snapshotting and selection are extremely fast operations.
 1. at eviction time, a snapshot of all the senders in the cache (sorted in _ascending_ score order) is created. This snapshot is then iterated over and senders are discarded (along with their transactions) as long as the high-load condition holds. Again, snapshotting and eviction are extremely fast.

### Cache capacity

The `TxCache` is `N` times larger than the others. Given `N = 5` and the configuration:

```
[TxDataPool]
    Size = 900000
    SizeInBytes = 524288000
    ...
```

it follows that the first cache has a capacity of `500000` transactions (or `~277 MB`) while the rest have a capacity of `100000` transactions (or `~56 MB`) each.

When capacity is close to be reached, **eviction** is triggered.

### Insertion in `TxCache`

 1. The eviction condition is tested before the actual addition
 1. If eviction is necessary, it is executed synchronously
 1. The incoming transaction is added in the cache if missing
 1. If the maximum capacity allocated for the sender is reached (currently, this is configured to be very high), a number of high-nonce transactions (of the sender in question) are removed from the cache so that the load (per sender) stays under the threshold.

### Selection of transactions from `TxCache`

The selection is invoked by the processing components. Typically, the *selection buffer* has a size of `numRequested = 30000` transactions and the sender-scoped batch size, is `batchSizePerSender = 10`.

 1. Once started, the selection stops when the *buffer* is full or when there are no more transactions to select
 1. Selection is performed in a loop, and its iterations are called **selection passes**
 1. At each *selection pass*, the senders are taken in **an approximate order** from the highest **score** to the lowest, and given the opportunity to contribute transactions. One bulk of transactions selected within a *pass* is called a **contribution**. Since there are *many selection passes*, a given sender can have *many contributions* during a single *selection*.
 1. One sender can contribute at most `batchSizePerSender * (score + 1)` in one selection pass. For example, a sender with `score == 100` could have a *contribution* of `10 * 101 = 1010` transactions is one selection pass.

#### Contributing transactions

In each *selection pass*, a sender may contribute transactions. The contribution itself depends on several factors.

 1. Transactions are contributed in the **nonce** order
 1. In the first *selection pass*, the sender is checked for an **initial nonce gap**
 1. If there is an *initial nonce gap*, the sender is blocked any contribution in the current selection, and the sender is tagged with `failedSelection`. All subsequent *selection passes* in the current selection will ignore the sender
 1. If the sender is tagged with `failedSelection` several (`5`) times in a row, it enters the **grace period**
 1. The *grace period* expires on its `3rd` `failedSelection` in a row
 1. During the *grace period* (currently defined as the `2nd` failed selection), the sender is permitted a contribution of *one transaction* per selection (only one selection since the grace period currently has a length of 1), which is called the **grace transaction**
 1. After the **grace period**, a `failedSelection` would result into marking the sender as **sweepable**
 1. A **sweepable** sender will is immediately removed (asynchronously) at the end of the selection
 1. Once the *initial nonce gap* is resolved, the `failedSelection` tag is removed (untagging happens at the very next selection)
 1. If, when contributing transactions, a *nonce gap* (called a *middle nonce gap*) is encountered, the sender is blocked any contribution in the current selection

### Score of senders in `TxCache`

The score for a sender is defined as follows:


```

                           (PPUAvg / PPUMin)^3
 rawScore = ------------------------------------------------
            [ln(txCount^2 + 1) + 1] * [ln(txSize^2 + 1) + 1]

                              1
 asymptoticScore = [(------------------) - 0.5] * 2
                     1 + exp(-rawScore)

```

For `asymptoticScore`, the [logistic function](https://en.wikipedia.org/wiki/Logistic_function) is used.

Notation:

 - `PPUAvg`: average gas points (fee) per processing unit, in micro ERD
 - `PPUMin`: minimum gas points (fee) per processing unit (given by economics.toml), in micro ERD
 - `txCount`: number of transactions
 - `txSize`: size of transactions, in kB (1000 bytes)

### Eviction

1. `TxCache` evicts a number (`TxPoolNumSendersToPreemptivelyEvict = 100`) of senders with low score when capacity is closed to be reached.
1. `TxCache` evicts high-nonce transactions of a sender when the maximum allowed quota for this sender is reached.
1. `CrossTxCache` evicts a number (`TxPoolNumTxsToPreemptivelyEvict = 1000`) of least-recently added transactions when capacity is reached. **The high-load capacity condition is checked per chunk** (as opposed to globally). But since distribution of items among the chunks is close to uniform, and the chunks are large, the eviction is reasonably efficient, reasonably rare (though generally a little bit greedier than an eviction with a globally checked high-load condition).
1. `CrossTxCache` does not evict **immune** items.
1. If `CrossTxCache` reaches its capacity (as stated, per chunk) but all items are **immune**, then eviction does not happen, addition does not happen; incoming item is simply discarded. This doesn't often happen in practice.
