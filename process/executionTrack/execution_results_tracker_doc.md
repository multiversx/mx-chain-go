# executionResultsTracker – Documentation

---

## 1. High‑Level Purpose

The `executionResultsTracker` keeps track of **ExecutionResults that have been produced but not yet notarized/confirmed**. It enforces **strict nonce continuity** (no gaps, no reordering) between the last notarized result and subsequently received results. It also supports **cleaning** (removing) results that have become confirmed by comparing them against a trusted header.

In short, the tracker:

- Remembers the **last notarized (confirmed) execution result**.
- Accepts and stores **new, sequential execution results** (pending confirmation).
- Lets callers query by **hash** or **nonce**.
- Produces the **ordered list of pending results**.
- **Cleans** results that are now confirmed, detecting three cases:
  - All match → OK (confirmed and removed).
  - Missing pending result(s) when compared to header → NotFound.
  - Mismatch between local pending vs header → Mismatch (cleans divergent tail).

Thread safety is provided using an internal `sync.RWMutex`.

---

## 2. Data Model & Fields

```go
type executionResultsTracker struct {
    lastNotarizedResult    *block.ExecutionResult
    mutex                  sync.RWMutex
    executionResultsByHash map[string]*block.ExecutionResult
    nonceHash              *nonceHash
    lastExecutedResultHash []byte
}
```

**Field meanings:**

- **lastNotarizedResult** – The most recent *confirmed* execution result acknowledged by consensus.
- **executionResultsByHash** – Map of **pending** (not yet notarized) execution results keyed by their `HeaderHash` (stringified).
- **nonceHash** – Helper index mapping `nonce -> hash` so we can look up pending results by nonce.
- **lastExecutedResultHash** – Tracks the hash of the *most recently added* execution result (pending or after cleaning). Used to quickly fetch the latest result in nonce sequence.

**Invariants (intended):**

1. `lastNotarizedResult` is non‑nil after initialization via `SetLastNotarizedResult`.
2. All entries in `executionResultsByHash` represent results with `Nonce > lastNotarizedResult.Nonce`.
3. Nonces in pending storage are **contiguous** starting at `lastNotarizedResult.Nonce + 1`.
4. `lastExecutedResultHash` always points to the highest nonce known (either last notarized, or the tip of pending results).

---

## 3. Lifecycle Overview

Typical usage pattern:

1. **Construct tracker** via `NewExecutionResultsTracker()`.
2. **Set initial confirmed state** with `SetLastNotarizedResult(...)` before accepting new results.
3. As new execution results arrive (in strict increasing nonce order): call `AddExecutionResult()`.
4. When a new header is notarized/confirmed (contains a sequence of execution results that should now be accepted as canonical): call `CleanConfirmedExecutionResults(header)`.
5. Use query helpers (`GetPendingExecutionResults`, `GetExecutionResultByHash`, `GetExecutionResultsByNonce`, `GetLastNotarizedExecutionResult`).

---

## 4. Status Types Returned by Clean

```go
type CleanResult string

const (
    CleanResultOK       CleanResult = "ok"
    CleanResultNotFound CleanResult = "result not found"
    CleanResultMismatch CleanResult = "different result"
)

// CleanInfo holds the result of a clean operation and the last matching nonce.
type CleanInfo struct {
    CleanResult             CleanResult
    LastMatchingResultNonce uint64
}
```

**Semantics:**

- **OK** – All header results matched tracker’s pending results (up to header length). Matching ones were removed from pending storage and `lastNotarizedResult` advanced.
- **NotFound** – Header contained more results than the tracker had pending; cleaning stopped at last match. `lastNotarizedResultHash` set to last matching hash. No mismatch was detected—just missing data.
- **Mismatch** – A pending result differed from the header at some index; all *remaining* pending results from that index forward were dropped (since local branch diverged). The tracker rolls back its executed tip to the last matching result.

---


