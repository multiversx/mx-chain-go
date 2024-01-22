# Sovereign Shard Local Setup Guide

This guide provides instructions for setting up a local sovereign shard, which is a fully independent chain with
capabilities akin to the MultiversX main chain. This includes smart contract processing, ESDT transfers, delegation,
staking, governance, guardians, and more.

## Prerequisites

- The scripts for bootstrapping a local sovereign chain are located in `mx-chain-go/scripts/testnet` and are currently
  available on the `feat/chain-go-sdk` branch.
- This tutorial is compatible with Ubuntu systems only.
- The epoch change concept is not available in this version, so any feature configuration that needs to be used should
  be set for epoch = 0.

## Sovereign Chain Architecture

Each sovereign shard requires two additional services for its operation:

1. **Notifier Service**: An observer on the main chain for a specific shard(in which your sc bridge is deployed). Set up
   instructions for the observer can be found [here](https://github.com/multiversx/mx-chain-observing-squad). Since the
   observer functions as the event notifier for the sovereign chain, it's important to enable the data export
   feature before starting it. To enable data export for the observer service, first locate the `external.toml` file
   within the observer's directory, specifically in `cmd/node/config`. Then, update the configuration to enable the
   feature by setting it to `true`::

```toml
[[HostDriversConfig]]
   # This flag shall only be used for observer nodes
   Enabled = false
```

2. **Bridge Transaction Sender Service**: Responsible for sending bridge transactions from the sovereign shard to the
   main chain. Validators in the sovereign shard collect all outgoing bridge transactions that need to be transferred to
   the main chain at the end of the round. The leader will collect necessary signatures from validators, aggregate them,
   and send an outgoing bridge transaction data to the service. This service will create a transaction with a hot wallet
   on the main chain to send to the bridge smart contract address.

For validators to collect data that needs to be bridged, they need to know the bridge smart contract deployed on their
sovereign shard. This can be specified in `cmd/sovereignnode/config/sovereignConfig.toml`
under `[OutgoingSubscribedEvents]`:

```toml
[OutgoingSubscribedEvents]
   SubscribedEvents = [
       { Identifier = "deposit", Addresses = ["your_bridge_sc_address"] }
   ]
```

Replace the address in the example with the deployed smart contract address.

The Bridge Transaction Sender Service holds sensitive data (your hot wallet on the mainchain), so secure communication
between the service and sovereign nodes is ensured through TLS certificates. When deploying your sovereign shard, the
scripts will generate a TLS certificate with its associated private key inside `~MultiversX/testnet/node/config`. There
you will find `certificate.crt` and `private_key.pem`, which are used for each sovereign node to communicate securely
with the bridge sender service.

### Setting Up Bridge Transaction Sender Service

1. Clone [the sovereign bridge repository](https://github.com/multiversx/mx-chain-sovereign-bridge-go)
2. Navigate to `server/cmd/server`.
3. Copy the TLS certificate (`certificate.crt`) and its associated private key (`private_key.pem`) from
   `~MultiversX/testnet/node/config`(after finishing the steps from `Node Configuration` chapter below).
4. Configure the service in the .env file:

```dotenv
# Multiversx main chain wallet to send bridge transactions.
WALLET_PATH="wallet.pem"
# Wallet's password (e.g.: json password encrypted wallet).
WALLET_PASSWORD=""
# MultiversX proxy (e.g.: https://testnet-gateway.multiversx.com)
MULTIVERSX_PROXY="https://testnet-gateway.multiversx.com"
# Bridge address on MultiversX to send transactions to
BRIDGE_SC_ADDRESS="your_bridge_sc_address"
# ....
```

4. Once you have configured all the necessary files for the bridge sender service, the next step is to build and run the
   application:

```bash
go build
```

5. After successfully building the application, you can start the service by running the binary:

```bash
./server
```

## Setup Configuration

### Nodes configuration

Inside `variables.sh` (from `mx-chain-go/scripts/testnet`), you can set up your node's configuration:

```bash
export SHARD_VALIDATORCOUNT=2
export SHARD_OBSERVERCOUNT=1
export SHARD_CONSENSUS_SIZE=2
```

### Sovereign chain connection with notifier

Once the observer is operational, it will begin exporting, sending data, and notifying your local sovereign chain. To
establish this connection, configure your sovereign chain to communicate with the observer. This is done in
the `cmd/sovereignnode/config/sovereignConfig.toml` file under the `[NotifierConfig.WebSocket]` section:

```toml
[NotifierConfig.WebSocket]
   Url = "localhost:22111"
```

If the observer is running on the same machine as your sovereign shard, the localhost URL as shown above will suffice.
However, if the observer is operating on a different machine, you should replace localhost with the appropriate IP
address.

Within the same configuration file (`sovereignConfig.toml`), there is another significant setting
under `[MainChainNotarization]`. This setting specifies the starting round from which all sovereign chain nodes should
begin notarizing main chain headers:

```toml
[MainChainNotarization]
   # This defines the starting round from which all sovereign chain nodes should start notarizing main chain headers
   MainChainNotarizationStartRound = 11
```

Additionally, after deploying your bridge smart contract on the mainnet, update the sovereign configuration file with
the smart contract address for the required subscribed events. This can be found under the `[NotifierConfig]` section:

```toml
[NotifierConfig]
SubscribedEvents = [
    { Identifier = "deposit", Addresses = ["your_bridge_sc_address"] },
    { Identifier = "executedBridgeOp", Addresses = ["your_bridge_sc_address"] }
]
```

Replace `your_bridge_sc_address` with the actual address of your deployed bridge smart contract.

### Gas Limit and Transaction Fees Configuration

Developers may wish to experiment with the gas limit per block, which can increase the number of transactions that can
be executed in the same block or reduce block time. These configurations, along with transaction fee settings, can be
found in `cmd/sovereignnode/config/economics.toml`. Modify these settings according to your specific requirements or
experimentation goals:

```toml
[FeeSettings]
   GasLimitSettings = [
       {EnableEpoch = 0, MaxGasLimitPerBlock = "1500000000", ...
```