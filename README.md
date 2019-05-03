# Elrond go sandbox

The go implementation for the Elrond Network testnet

# Getting started

### Prerequisites

Building the repository requires Go (version 1.12 or later)

### Installation and running

Run in  %project_folder%/cmd/seednode folder the following command to build a seednode (first node in the network
 used for bootstrapping the nodes):
 
 ```
 $ go build
 $ ./seednode
 ```
 
Run in  %project_folder%/cmd/node folder the following command to build a node:

```
$ go build
$ ./node -port 23000 -private-key "b5671723b8c64b16b3d4f5a2db9a2e3b61426e87c945b5453279f0701a10c70f"
```

For the network customization please take a look in the p2p.toml<br />
For the consensus group customization please take a look in the /cmd/node/genesis.json<br />
To run multiple nodes take a look at the end of this document.

### Running the tests
```
$ go test ./...
```

# Progress
### Done
- [x] Cryptography
  - [x] Schnorr Signature
  - [x] Belare-Neven Signature
  - [x] BLS Signature
  - [x] Modified BLS Multi-signature
- [x] Datastructures
  - [x] Transaction
  - [x] Block
  - [x] Account
  - [x] Trie
- [x] Execution
  - [x] Transaction
  - [x] Block
  - [x] State update
  - [x] Synchronization
  - [x] Shard Fork choice
- [x] Peer2Peer - libp2p
- [x] Consensus - SPoS
- [x] Sharding - fixed number
  - [x] Transaction dispatcher 
  - [x] Transaction
  - [x] State
- [x] MetaChain
  - [x] Data Structures
  - [x] Block Processor
  - [x] Interceptors/Resolvers
- [x] VM - K-Framework
  - [x] K Framework go backend
  - [x] IELE Core
  - [x] IELE Core tests
- [x] Smart Contracts on a Sharded Architecture
  - [x] Concept reviewed
- [x] Governance
  - [x] Concept reviewed
- [x] Testing 
  - [x] Unit tests
  - [x] Integration tests
  - [x] TeamCity continuous integration
  - [x] Manual testing

### In progress
- [ ] Sharding - fixed number
  - [ ] Nodes dispatcher (shuffling)
  - [ ] Network
- [ ] MetaChain Consensus
- [ ] VM - K-Framework
  - [ ] IELE Adapter
  - [ ] EVM Core
  - [ ] EVM Core tests
  - [ ] EVM Adapter
- [ ] Smart Contracts on a Sharded Architecture
  - [ ] VM integration
  - [ ] SC Deployment
  - [ ] Dependency checker + SC migration
  - [ ] Storage rent + SC backup & restore
  - [ ] Request-response fallback
- [ ] Fee structure
- [ ] Governance
  - [ ] SC for ERD IP
  - [ ] Enforced Upgrade mechanism for voted ERD IP
- [ ] Testing
  - [ ] Automate tests with AWS 
- [ ] Bugfixing

### Backlog
- [ ] Adaptive State Sharding
  - [ ] Splitting
  - [ ] Merging 
  - [ ] Redundancy
- [ ] Privacy
- [ ] DEX integration
- [ ] Interoperability
- [ ] Optimizations
  - [ ] Randomness
  - [ ] Consensus
  - [ ] Smart Contract 

# Private/Public Keys for testing

```
# Start n instances in a group of <consensusGroupSize> (genesis.json) from a total of <initialNodes> (genesis.json) validators,<br />
# each of them having maximum -max-allowed-peers (flag) peers, with round duration of <roundDuration> (genesis.json) miliseconds<br />

# Private keys:

#  1: b5671723b8c64b16b3d4f5a2db9a2e3b61426e87c945b5453279f0701a10c70f
#  2: 8c7e9c79206f2bf7425050dc14d9b220596cee91c09bcbdd1579297572b63109
#  3: 52964f3887b72ea0bd385ee3b60d2321034256d17cb4eec8333d4a4ce1692b08
#  4: 2c5a2b1d724c21be3aebf46dfa1db841b8b58d063066c19e983ff03b3f955f08
#  5: 6532ccdb32e95ca3f4e5fc5b43c41dda610a469f7f18c76c278b9b559779300a
#  6: 772f371cafb44da6ade4af11c5799bd1c25bbdfb17335f4fc102a81b2d66cc04
#  7: 12ffff943b39b21f1c5f1455e06a2ab60d442ff9cb65451334551a0e84049409
#  8: a7160d033389e99198331a4c9e3c7417722ecc29246f42049335e972e4df5b0f
#  9: 9cf7b345fdf3c6d2de2d6b28cc0019c02966ef88774069d530b636f760292c00
# 10: f236b2f60ad8864ea89fd69bf74ec65f64bd82c2f310b81a8492ba93e8b6c402
# 11: 0f04b269d382944c5c246264816567c9a33b2a9bf78f075d9a17b13e7b925603
# 12: 8cf6e6aeb878ef01399e413bc7dd788a69221a37c29021bd3851f2f5fe67f203
# 13: c7f48a69e4b2159fe209bdb4608410516f28186ad498ca78b16d8b2bebfb1f0f
# 14: 7579d506ff015e5e720b2e75e784c13a4662f48b6e2038af6e902b1157239101
# 15: b7877c28e394ab4c89d80e8b2818ef1346ee8c0fdd6566a6d27088ad097e4f05
# 16: 055ae06aad2c7f8d50ecd4bd7c4145cb19636b0b0126ffa4ee1326afb3876000
# 17: c47b89db3e3ad067863af5a7b7f9e9dec0e47516e87d5d6d744e0af581a79404
# 18: 843c4bea60b629fae50a0334ba9c7284f886b90502b740c8f95ab13a36a08c0e
# 19: 92561fd546014adcd13ff7776829f1c8c0886e83eb04fb723fc3636da8f2960b
# 20: 22a3922963cc1fe57a59178f021282223a8742fb4476f7a5c5b4c2c2aa2d4f0f
# 21: 02c9d56e503857832c07d78b0d75aabb8e6c109e9cec641b8681afaee2c9a701

# Public keys:

#  1: 5126b6505a73e59a994caa8f556f8c335d4399229de42102bb4814ca261c7419
#  2: 8e0b815be8026a6732eea132e113913c12e2f5b19f25e86a8403afffbaf02088
#  3: e6ec171959063bd0d61f95a52de73d6a16e649ba5fa8b12663b092b48cc99434
#  4: 20ccf92c80065a0f9ce1f1b8e3dee1e0b9774f4eebf2af7e8fa3ac503923360d
#  5: 9a3b8e67f42aef9544e0888ea9daee77af90292c86336203d224691d55306d08
#  6: 0740bccedc28084ab811065cb618fec4ee623384b4b3d5466190d11ff6d77007
#  7: 0ccba0f98829ea9f337035a1f7b13cbd8e9ffb94f2c538e2cafb34ca7f2bcd24
#  8: d9e9596c28a3945253d46bc1b9418963c0672a26a0b40ee7372cb9ec34d1ee07
#  9: 86fbd8606e73b7a4f45a51b443270f3050aff571a29b9804d2444c081560d1dd
# 10: 2084f2493e68443a5b156ec42a8cd9072c47aa453df4acd20524792a4fd9f474
# 11: f91d24256d918144aaacfa641cd113af05d56cfb7a5b8ba5885ebd8edd43fe1e
# 12: e8d4bcfe91c3c7788d8ab3704b192229900ec3fe3f1eb6f841c440e223d401a0
# 13: 4bf7ee0e17a0b76d3837494d3950113d3e77db055b2c07c9cb443f529d73c8e3
# 14: 20f12f7bdd4ab65321eb58ce8f90eec733e3e9a4cc9d6d5d7e57d2e86c6c2c76
# 15: 34cf226f4d62a22e4993a1a2835f05a4bb2fb48304e16f2dc18f99b39c496f7d
# 16: b9f0fc3e1baa49c027205946af7d6c79b749481e5ab766356db3b878c0929558
# 17: 6670b048a3f9d93fdacb4d60ff7c2f3bd7440d5175ca8b9d2475a444cd7a129b
# 18: d82b3f4490ccb2ffbba5695c1b7c345a5709584737a263999c77cc1a09136de1
# 19: 29ba49f47e2b86b143418db31c696791215236925802ea1f219780e360a8209e
# 20: 199866d09b8385023c25f261460d4d20ae0d5bc72ddf1fa5c1b32768167a8fb0
# 21: 0098f7634d7327139848a0f6ad926051596e5a0f692adfb671ab02092b77181d


gnome-terminal -- ./node -port 23000 -private-key "b5671723b8c64b16b3d4f5a2db9a2e3b61426e87c945b5453279f0701a10c70f"
gnome-terminal -- ./node -port 23001 -private-key "8c7e9c79206f2bf7425050dc14d9b220596cee91c09bcbdd1579297572b63109"
gnome-terminal -- ./node -port 23002 -private-key "52964f3887b72ea0bd385ee3b60d2321034256d17cb4eec8333d4a4ce1692b08"
```
