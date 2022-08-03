# Contract sources here

We should be striving to document here where all the contract source codes lie.

## execute-on-dest-esdt-issue-*.wasm

Files:
    execute-on-dest-esdt-issue-parent-0.34.1.wasm
    execute-on-dest-esdt-issue-child-0.34.1.wasm
Repo: https://github.com/ElrondNetwork/elrond-wasm-rs
Commit: 0947f9c3e1c942ee165853fcb8d50afcecdf938a
Paths:
    contracts/feature-tests/composability/execute-on-dest-esdt-issue-callback/parent
    contracts/feature-tests/composability/execute-on-dest-esdt-issue-callback/child

## forwarder-raw-0.34.0.wasm

All it does is send transactions to other contracts or wallets.

Updated slightly after the release of elrond-wasm 0.34.0, might rename to a future release, to be easier to find.

Repo: https://github.com/ElrondNetwork/elrond-wasm-rs
Commit: 0947f9c3e1c942ee165853fcb8d50afcecdf938a
Path: contracts/feature-tests/composability/forwarder-raw/src/forwarder_raw.rs
Quick link: https://github.com/ElrondNetwork/elrond-wasm-rs/blob/0947f9c3e1c942ee165853fcb8d50afcecdf938a/contracts/feature-tests/composability/forwarder-raw/src/forwarder_raw.rs

## vault-0.34.0.wasm

Receives payments and will send EGLD or tokens back on request.

Updated slightly after the release of elrond-wasm 0.34.0, might rename to a future release, to be easier to find.

Repo: https://github.com/ElrondNetwork/elrond-wasm-rs
Commit: 0947f9c3e1c942ee165853fcb8d50afcecdf938a
Quick link: https://github.com/ElrondNetwork/elrond-wasm-rs/blob/0947f9c3e1c942ee165853fcb8d50afcecdf938a/contracts/feature-tests/composability/vault/src/vault.rs

## use-module-0.34.1.wasm

Tests various standard modules.

Repo: https://github.com/ElrondNetwork/elrond-wasm-rs
Commit: feddb5b7ea5b0b5bb0f2d5a2ead65797252e3606
