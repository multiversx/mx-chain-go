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

Once you finished testing, you can close the sovereign notifier (can use CTRL+C) and both the local testnet and the
sovereign shard, by executing:

```bash
cd scripts
./local-testnet.sh stop
./sovereign-shard.sh stop
```

## Alternative demo (lighter version)

To set up a fully customized environment for a sovereign shard receiving incoming transactions on a local testnet/main
chain, there are several prerequisites involved, such as deploying smart contracts, issuing tokens, and performing
transfers. It's worth noting that running two local chains and a notifier on the same machine can be resource-intensive.

To simplify the testing process for incoming ESDT/NFT transfers on the sovereign shard, you can utilize the provided
"mockNotifier" tool. This tool functions as an observer notifier specifically designed for a sovereign shard. It
triggers the transmission of blocks, starting from an arbitrary nonce, with incoming events occurring every 3 blocks.
Each event includes the transfer of an NFT and an ESDT token. The periodic transmission consists of 2 NFTs
(ASH-a642d1-01 & ASH-a642d1-02) and one ESDT token (WEGLD-bd4d79).

To verify the success of token transfers, you can utilize the sovereign proxy for the subscribed address, which can be
found in the [sovereignConfig.toml](../../sovereignnode/config/sovereignConfig.toml) file. By using the following API
endpoint: `http://127.0.0.1:7950/address/:subscribed-address/esdt`, you can check the status of the token transfers.
It's important to note that the blocks are sent with an arbitrary period between them, allowing for flexibility in the
testing process.

### How to use

1. Go to [scripts](../../../scripts/testnet)
2. Prepare and start the sovereign chain

```bash
./prerequisites.sh 
./clean.sh
./config.sh
./sovereignStart.sh
```

3. Go to [mock notifier](mockNotifier) and start it:

```bash
 go build
 ./mockNotifier
```

Once started, one should see a similar log:

```
INFO [2023-06-23 12:35:06.682]   wsServer.initializeServer(): initializing WebSocket server url = localhost:22111 path = /save 
INFO [2023-06-23 12:35:06.682]   sending block                            nonce = 10 hash = 81c7e3552bab0b475c330642dbc17c91df23afeba72700a3376e32f2cc05c1fc prev hash = 70d6e2419a3a83de44a5f59cc4725ae1fd3b697d85e0fd0e59a74b338eea797a rand seed = 962657265827606ec30140fd770292f0bd9d403c640bdac8fcaa7babe98d4930 prev rand seed = 2d04a4641ebd0ce0f0d5852c44507b5737d08574a5aa31d04e9843348a37a819 
WARN [2023-06-23 12:35:06.682]   could not send data                      topic = SaveBlock error = no client connected 
INFO [2023-06-23 12:35:08.730]   new connection                           route = /save remote address = 127.0.0.1:35446 
INFO [2023-06-23 12:35:08.811]   new connection                           route = /save remote address = 127.0.0.1:35454 
INFO [2023-06-23 12:35:09.489]   new connection                           route = /save remote address = 127.0.0.1:35466 
INFO [2023-06-23 12:35:16.813]   sending block                            nonce = 11 hash = 39c0eac07606eb13e3361ceca67b9ed6f3c4ac2fdfe80b58fe6defd19cb4162c prev hash = 81c7e3552bab0b475c330642dbc17c91df23afeba72700a3376e32f2cc05c1fc rand seed = 1c8c74f9f9057c363807bb2d7f37547f06419a3a8c5aec280934ebfbd5df7074 prev rand seed = 962657265827606ec30140fd770292f0bd9d403c640bdac8fcaa7babe98d4930 
INFO [2023-06-23 12:35:19.815]   sending block                            nonce = 12 hash = de9edd62acebc22509267589cbf31e4fed519b5705085bb94a0fe5615262bf78 prev hash = 39c0eac07606eb13e3361ceca67b9ed6f3c4ac2fdfe80b58fe6defd19cb4162c rand seed = d1d35aee2e74e6a0674d8d158b11da2475c67c1280d80af489d28f15f4e2c4af prev rand seed = 1c8c74f9f9057c363807bb2d7f37547f06419a3a8c5aec280934ebfbd5df7074 

```