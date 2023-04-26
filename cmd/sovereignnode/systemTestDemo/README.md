## About

This folder contains system tests scripts for a sovereign chain that is receiving incoming txs notifications from a
local testnet.

### How it works

The scripts found in `scripts` folder:

- `local-testnet.sh` creates a local testnet (main chain) with one shard + metachain (each having 1 validator + 1
  observer)
- `sovereign-notifier.sh` creates a sovereign notifier node (which is a local testnet observer started in shard 0).
  This node acts a server outport driver for sovereign chain. It will send block data information about incoming blocks
  from the main chain to sovereign chain.
- `sovereign-shard.sh` creates sovereign shard with 2 validators + 1 observer. All nodes from the sovereign chain will
  be notified about incoming blocks received from the notifier above.

Once the setup is ready, the sovereign shard will start receiving finalized main chain blocks. One can see a similar
log info:

```
INFO [2023-04-25 17:01:22.025] [wsClient]           [0/0/288/(END_ROUND)] processing payload                       counter = 984 operation type = FinalizedBlock message length = 61
INFO [2023-04-25 17:01:22.025] [..overeign-process] [0/0/288/(END_ROUND)] notifying shard extended header          hash = f1331c8f759c9e341c08a7a430e2c2e643433d994121f4cf9582484d4e92afd4
```

Besides blocks, sovereign shard is also subscribed to receive notification about incoming transactions(specified by its
subscribed addresses). Every time a new transaction coming from main chain to sovereign chain is finalized, a similar
log can be found:

```
INFO [2023-04-24 17:08:42.024] [..overeign-process] [0/0/106/(END_ROUND)] found incoming tx                        tx hash = d0f77769ef1284c8f7c570278f33ceb823cee73949507b9c6fe2d62b3eae85c5 receiver = 000000000000000005005ebeb3515cb42056a81d42adaf756a3f63a360bfb055 
```

Inside this folder, there is a `main.go` script which sends a transaction to one of the subscribed sovereign addresses.
The address is specified by `SubscribedAddresses` field from this config
file `mx-chain-go/cmd/sovereignnode/config/notifierConfig.toml`

## How to use

1. Create and start a local testnet/main chain:

```bash
cd scripts
./local-testnet.sh new
./local-testnet.sh start
```

2. Start the sovereign notifier:

```bash
./sovereign-notifier.sh
```

3. Create and start a sovereign chain:

```bash
./sovereign-shard.sh new
./sovereign-shard.sh start
```

If you want to send transactions from main chain -> sovereign chain, inside `systemTestDemo` folder, you need to:
```bash
go build
./sovereignChecker
```
Once transaction is executed, one should see a similar log:
```
INFO [2023-04-26 14:03:03.196]   loaded                                   address = erd1s3uetuj3g86fx062j9tk9pdsmet7vksrk059xj86drk639dfenys728gme file = scripts/testnet/testnet-local/sandbox/proxy/config/walletKey.pem 
INFO [2023-04-26 14:03:03.199]   sent transaction                         tx hash = cb23013a0117c6d05d0a27faedcee01fd56e2821197208f6dffa8af87b1a4231 
INFO [2023-04-26 14:03:03.199]   waiting 5 rounds...                      
INFO [2023-04-26 14:03:28.237]   finished successfully                    tx result from api = {"data":{"tran...
```