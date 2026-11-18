# Promise callback fixture

`promise-probe.wat` is the source for `promise-probe.wasm`. Tests deploy the
checked-in Wasm directly; no compiler or Python installation is required to run them.
The receiver is the existing `integrationTests/vm/txsFee/testdata/forwarderQueue/vault-promises.wasm`.

The `start` endpoint accepts a destination, endpoint name, remote gas budget and
callback mode: 0 succeeds, 1 writes then signals an error, 2 writes then exhausts
its metered gas in a finite loop. Successful callbacks increment a stored counter;
the other two modes must roll back that increment. This is a test fixture, not an
application contract: exported callbacks deliberately have no caller authorization.

The `send(destination)` endpoint forwards the incoming EGLD as a pure transfer,
without a destination endpoint, for the contract payability matrix.

Rebuild with a WAT compiler, for example:

```sh
wat2wasm promise-probe.wat -o promise-probe.wasm
```

The included binary was built with Wasmtime 48.0.0's `wat2wasm` compiler.
Compiler-specific custom name sections can change the file hash without changing
execution. Keep the source and binary together and rerun the Go promise and payability tests after rebuilding.
