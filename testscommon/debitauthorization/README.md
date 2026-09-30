# Signed EGLD debit invariant

This package implements a **test observer**, installed before constructing the
node transaction processor and the VM blockchain hook. It does not add a new
consensus rule. The initial profile covers native EGLD in ordinary accounts,
ordinary transactions (including guarded accounts), and relayed-v3 transactions.
Smart-contract addresses are classified using `core.IsSmartContractAddress` and
excluded from the spending invariant.

For each observed execution and ordinary account, the observer checks:

1. Every saved, non-reverted debit has an applicable verified transaction signer.
2. Gross debits do not exceed that account's signed value and fee ceilings.
3. Credits, including refunds, do not replenish the spending ceiling.
4. Initial balance plus saved credits minus saved debits equals actual state.
5. The transaction executed is unchanged from the transaction verified.

For ordinary transactions the sender authorizes value and fees. For relayed v3,
the sender authorizes value and the relayer authorizes fees. The fee ceiling is
the signed gas limit multiplied by the signed gas price using arbitrary-precision
arithmetic. It is a ceiling, not an independently calculated exact transaction
fee. Legacy relayed v1/v2 envelopes are explicitly rejected as unsupported rather
than being interpreted with v3 funding rules.

## Use

Create the fixture with `vm.CreatePreparedTxProcessorWithAccountsDecorator` and
install `debitauthorization.New(accounts)` in the decorator. All node and VM
consumers must use this adapter. Initialize and commit fixture state before
observing execution.

Call `observer.Begin(tx, verify)` and execute **the returned transaction copy**.
The verifier must check the real signatures and the applicable chain, version,
activation and guardian rules. Begin also checks the sender's execution nonce.
Call `observer.Finish()` after transaction processing and require a nil error
before committing the accounts. Also assert return codes, intended recipients,
exact value/fee effects and applicable roots from independent fixture expectations.

`integrationTests/vm/txsFee/signedDebit_test.go` demonstrates production signature
verification with real Ed25519 keys, a cache that always misses, the stateful
guardian policy, and a real accounts adapter and VM. It exercises ordinary and
relayed-v3 transfers, invalid signatures, signed-field changes, chain/nonce checks,
guarded transfers, fee-paying failed execution, a normal WASM call, and state
rollback followed by re-execution. Contract deployment in the call test is fixture
setup outside the observed interval.

## Observation boundaries

The account wrapper observes both balance methods, including negative additions,
and associates changes with the real journal index at SaveAccount. Finish counts
saved events even when a later credit makes the net balance increase. Unsaved
changes are not committed debits. Successful RevertToSnapshot calls mark undone
events; restoration does not become a new spending debit. Reacquire account
handles after a revert, as for the underlying accounts adapter.

Finish must precede Commit/CommitInEpoch, which reset journal positions. It ends
the observation interval; fixtures must not insert additional mutations between
Finish and Commit. The report's ExecutionID is a SHA-256 digest of the copied
protobuf transaction for test correlation, not the network transaction hash.

The observer is for serial processing. Direct access to the underlying adapter,
trie recreation/state synchronization during an interval, unobserved mutation
interfaces, asynchronous continuations spanning intervals, cross-shard lineage,
legacy relayed funding, ESDT balances and full-block rollback of non-state
collectors are outside this profile. Saving an ordinary account without its
wrapper is reported as an instrumentation error.

It checks principal authorization and signed ceilings, not every destination or
purpose constraint. In particular, a sender's value and fee ceilings are combined
at the balance-method boundary; tests must independently check their allocation
and the exact resulting effects. A configured verifier or a passing observer
self-test alone is not evidence that signatures were verified by a real node.

## Validation

From the repository root:

```sh
make test-debit-authorization
go test -race -count=1 ./testscommon/debitauthorization
```

The seven signed integration tests also run under `go test -short ./...` and
`go test -short -race ./...`; their fixture no longer skips short mode. The dedicated
target selects them without running the rest of the repository. The coverage target
excludes integration-test packages, so it does not include this profile. Observer self-tests use explicitly
labeled verified fixtures and test the assertion mechanism, including a debit
masked by larger credits. They do not reproduce an underlying protocol exploit.

On the checked Go 1.26 / VM v1.5.49 combination, the normal WASM scenario passed,
but a full `-race` run aborted during its deployment fixture with a `checkptr`
diagnostic in the existing Wasmer2 CGO callback binding. The observer's race tests
and the transaction-only integration race profile are separate checks. The native
failure is not suppressed, and passing ordinary execution does not qualify the
native boundary under instrumentation.
