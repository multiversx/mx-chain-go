# stateAnalysis

Instrumentation for state and trie analysis during import-db re-execution.

The package holds one global collector. When `StateAnalysisConfig.Enabled` is false (the default) every
hook is a single atomic load and returns; no file is created. When enabled, counters are accumulated
with atomics by hooks in `trie`, `state`, `storage/pruning` and `process/block`, snapshotted at every
block end, aggregated into rows of `BlocksPerRow` blocks and written as gzipped csv under
`<OutputDir>/shard_<id>/`.

Enable it in `config.toml`:

```toml
[StateAnalysisConfig]
    Enabled = true
    Label = "master"          # tag stored in every row, use the branch name under test
    OutputDir = "state-analysis"
```

## Output files

| File | Row | Content |
| --- | --- | --- |
| `blocks.csv.gz` | every `BlocksPerRow` blocks (and at every epoch change) | sum of every counter over the blocks of the row, `max_<counter>` for timers and selected counters (maximum per-block value inside the row), running maxima reset at every block (`*_max`), and the last sampled gauges |
| `outliers.csv.gz` | one block | same columns as `blocks.csv.gz`, written for every block whose commit time, storage reads or journal length exceed the configured thresholds |
| `reuse.csv.gz` | per epoch and kind | reuse-distance histogram: `count_bN` / `bytes_bN` where `b0` = touched again in the same block, `bN` = touched again after `[2^(N-1), 2^N)` blocks, last bucket = never seen before (bounded by `MaxTrackedKeys`, `resets` counts map drops) |
| `hotset.csv.gz` | per epoch, kind and rank | the `TopKPerEpoch` most touched keys |
| `snapshots.csv.gz` | per state snapshot | duration, nodes, throttler waits, counts of main trie nodes, data tries, accounts with code / data trie |
| `structure/epoch_<N>_<unit>.json` | per state snapshot | node counts and bytes by type and depth, branch fan-out, extension and leaf key lengths, leaf value size distribution, leaf versions, estimated data-leaf metadata, estimated size under a compact encoding, data-trie size distribution, account and code entry counts |
| `profiles/` | sparse | CPU profile of the first `ProfileNumBlocks` blocks of every `ProfileEveryNEpochs`-th epoch and a heap profile at the end of that window |

Every row carries `shard`, `label`, `epoch`, the first round and the first and last nonce of the row.
Files are opened in append mode (header written only when the file is new), so a component restart
inside one process appends a new gzip member instead of truncating the previous rows.

## Column families

Suffixes: `_user` / `_peer` = accounts adapter; `_main` / `_data` / `_peer` = trie kind (user main trie,
user data tries, peer trie); `_accounts` / `_peer` / `_other` = storage unit; `_branch` / `_ext` /
`_leaf` = node type; `_bN` = histogram bucket; `_ns` = nanoseconds.

| Prefix | Hooked in | Answers |
| --- | --- | --- |
| `blk_` | `process/block` | wall, process, commit-state and CPU time per block, snapshot stalls, aborted blocks |
| `adb_` | `state/accountsDB.go` | account loads / saves / removes, save re-read cost, saves with unchanged bytes, repeated saves of one address per block, code loads, data-trie holder hits vs recreates, RootHash cost per save, journal length, reverts, trie recreates, commit split (data tries / main trie / eviction / state accesses), unique accounts and data tries per block and their overlap with the previous block |
| `dt_` | `state/trackableDataTrie` | dirty-map hits, hashed vs raw lookups and their hit rates, key hashes computed, dirty keys per save, inserts / modifies / deletes / no-op writes / migrations |
| `trie_` | `trie` | gets and found depth, updates, deletes, RootHash calls vs actual computations, commits with hash and write time, recreates, node loads by type and by depth in the Get path, resident hits, decode time, encode and hash calls, child re-encodes in `branchNode.hashNode`, cached hash short-circuits, dirty nodes committed by type, collapses at the max in-memory level |
| `st_` | `trie/trieStorageManager.go`, `storage/pruning` | storage manager gets and puts with bytes and time, pruning storer cache hits / persister index hits / misses, LevelDB get and put latency and bytes per unit, snapshot reads and writes |
| `prune_` | `state/storagePruningManager` | eviction waiting list puts, wipes and dropped hashes, pruning buffer overflows, nodes removed |
| `mem_`, `trie_resident_*`, `st_cache_*`, `adb_holder_*` | sampled every `MemorySampleEveryNBlocks` | heap, RSS, GC, resident trie nodes and bytes by type, storage cache size, data tries held at commit |

The trie kind of node-level counters is taken from a process-wide context set by the accounts adapter
and the trackable data trie around their trie calls. Import-db processes blocks on one goroutine, so
the attribution is exact there; concurrent API readers would be attributed to the current kind.
