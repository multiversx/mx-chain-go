# snapshotAnalysis

Measurements of the state snapshot process: what a snapshot does, where its goroutines spend their time,
how it interacts with block processing, storage and pruning, and which work is duplicated.

The package holds one global collector. When `SnapshotAnalysisConfig.Enabled` is false every
hook is an atomic load (or a nil check) and returns; no file is created. On this branch the shipped
`cmd/node/config/config.toml` enables it by default. When enabled, the hooks in
`state`, `trie`, `storage/pruning`, `statusHandler` and `process/block` feed one `Snapshot` object per
snapshot in progress (one per trie storage manager: `UserAccountsUnit`, `PeerAccountsUnit`) and write
gzipped csv files under `<OutputDir>/shard_<id>/`.

The section in `config.toml` (enabled by default on this branch; set `Enabled = false` to turn it off):

```toml
[SnapshotAnalysisConfig]
    Enabled = true
    Label = "master"              # tag stored in every row, use the branch name under test
    OutputDir = "snapshot-analysis"
    SampleIntervalMs = 1000       # timeline resolution
    AfterSnapshotSeconds = 120    # keep sampling after completion (pruning backlog, cache)
```

A summary of every snapshot is also logged at INFO level (`snapshot analysis summary`).

## How a snapshot runs (what is measured)

1. `SnapshotState` (epoch start commit) dedups the request, sets the in-progress flag, blocks pruning
   (`EnterPruningBufferingMode`) and launches `snapshotState`; in import-db (or with serialized snapshots)
   the caller then **blocks** until every trie traversal finished (`stall_ns`).
2. `snapshotState` waits for the trie storer to open the persister of the snapshot epoch (`wait_epoch_ns`),
   checks `ShouldTakeSnapshot` (`check_ns`) and starts three goroutines: the main trie `TakeSnapshot` plus the
   **leaves consumer**, the **missing nodes syncer** and the **completion** waiter.
3. `TakeSnapshot` enqueues a request on the trie storage manager queue (`enqueue_wait_ns` when the queue is
   full). The dispatcher goroutine waits for a free slot of the goroutine throttler (`throttler_wait_ns`,
   `SnapshotsGoroutinesPerCore * GOMAXPROCS` slots) and starts one **traversal goroutine** per trie
   (`queue_latency_ns` from enqueue to start).
4. A traversal goroutine does, for every node: wait while the node is busy processing a block (`parked_ns`,
   polled every 1 ms), `GetWithoutAddingToCache` (cache lookup, then the epoch persisters from the newest
   allowed one down: `get_*`, `get_found_epochN`), decode (`decode_ns`), recurse, then
   `PutInEpochWithoutCache` into the snapshot epoch (`put_*`). A node found in an epoch older than
   `epoch-1` is also written into `epoch-1` (`prev_epoch_puts`). A node found in the snapshot epoch itself
   is written again anyway (`put_redundant_same_epoch`). Main trie leaves are pushed on the leaves channel
   (`leaves_send_wait_ns` when the consumer is slower than the traversal).
5. The leaves consumer unmarshals every main trie leaf (`leaves_unmarshal_ns`, `leaves_recv_wait_ns` when it
   waits for the traversal) and enqueues one `TakeSnapshot` per account with a data trie. Identical data
   tries (same root hash) are traversed and written once per account (`dup_data_tries`, `dup_data_trie_nodes`).
6. The completion waiter waits for every traversal (`traversal_ns`, the tail of data tries after the main
   trie is `data_tries_tail_ns`), closes the missing nodes channel, waits for the syncer (`sync_wait_ns`),
   unblocks pruning, writes the `activeDB` marker (`completion_ns`).
7. After the snapshot, the first `PruneTrie` resolves every buffered pruning operation in one go
   (`prune_buffered`, `prune_resolve_*`, `prune_removed`), under the accounts adapter lock.

Storage level: the trie storer is a `triePruningStorer`. Snapshot reads hold its read lock while probing
the persisters (`st_snap_get_rlock_ns`, `st_snap_get_probes`); every snapshot write takes the storer
**write lock** (`createAndInitPersisterIfClosedProtected`, `st_snap_put_lock_ns`) before the serial
LevelDB `Put` (`st_snap_put_ns`, batched, a batch flush is a synchronous write that shows up in the
`gt100ms` buckets). Processing reads and writes of the same storer are reported as `st_proc_*`.

## Output files

| File | Row | Content |
| --- | --- | --- |
| `snapshots.csv.gz` | one per snapshot, written when its after-window ends | phases, node and byte counts, read / write / decode / parked times and latency histograms, goroutine utilization, channel and queue waits, duplicate work, missing nodes, processing activity alongside (busy time, commits, process CPU), storage counters (`st_*`), pruning backlog, heap |
| `timeline.csv.gz` | every `SampleIntervalMs` while a snapshot runs and for `AfterSnapshotSeconds` after it | the same quantities as deltas (`*_d`) plus gauges: active / parked goroutines, queue and channel lengths, node busy %, process CPU cores, heap, cache size, pruning buffer and eviction waiting list lengths. `phase` is one of `waiting_epoch`, `checking`, `starting`, `main_trie`, `data_tries`, `finishing`, `after`, `skipped` |
| `tries.csv.gz` | the main trie of every snapshot and every data trie with at least 1000 nodes | address, root hash, start offset, wall time, nodes, bytes, depth, get / put / decode / parked time |

All durations are nanoseconds (`_ns`); `_d` columns are deltas since the previous timeline row; `_le100us`,
`_le1ms`, `_le10ms`, `_le100ms`, `_gt100ms` are latency histogram buckets (counts).

## Derived quantities worth computing

- goroutine utilization = `goroutine_wall_ns / (traversal_ns * goroutine_max_active)`; the share of that
  wall time in I/O is `goroutine_io_ns / goroutine_wall_ns`, in decode `decode_ns / goroutine_wall_ns`,
  parked `parked_ns / goroutine_wall_ns`; what is left (`goroutine_other_ns`) is scheduling, channel sends
  and the work of the hooks themselves.
- read amplification = `st_snap_get_probes / (st_snap_gets - st_snap_cache_hits)`: persisters probed per read.
- lock pressure = `st_snap_put_lock_ns / st_snap_puts` and `st_snap_get_rlock_ns / st_snap_gets`.
- redundant writes = `put_redundant_same_epoch / puts`, duplicated tries = `dup_data_trie_nodes / data_nodes`,
  duplicated node writes (estimate, sampled) = `dup_node_ratio`.
- main trie vs data tries: `main_trie_ns` vs `data_tries_tail_ns`; a long tail with few active goroutines
  on the timeline means the last large data tries serialize the end of the snapshot.
- processing impact (live node): `node_busy_pct` on the timeline against `parked_ns_d`; commits during the
  snapshot (`commits`, `commit_ns`, `commit_max_ns`) against commits in the after-window.

## Overhead

Per node: a few `time.Now()` calls and local counter increments; counters are flushed to the shared
snapshot object every 1024 nodes. The sampled node hash set costs about `nodes / 2^NodeHashSampleShift * 40`
bytes per snapshot, the data trie root set about 40 bytes per distinct data trie (bounded by
`MaxTrackedDataTrieRoots`). Both are released when the snapshot row is written.
