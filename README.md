# Elrond go

The go implementation for the Elrond Network testnet

## Installation and running

In order to join the network as an observer or as a validator, the required steps are explained below:

### Step 1: install & configure go:
The installation of go should proceed as shown in official golang installation guide https://golang.org/doc/install . In order to run the node, minimum golang version should be 1.12.4.

### Step 2: clone the repository and build the binaries:
The main branch that will be used is the master branch. Alternatively, an older release tag can be used.

```
$ mkdir -p $GOPATH/src/github.com/ElrondNetwork
$ cd $GOPATH/src/github.com/ElrondNetwork
$ git clone https://github.com/ElrondNetwork/elrond-go-sandbox
$ cd elrond-go-sandbox && git checkout master
$ GO111MODULE=on go mod vendor
$ cd cmd/node && go build
```

### Step 3: creating the node’s identity:
In order to be registered in the Elrond Network, a node must possess 2 types of (secret key, public key) pairs. One is used to identify the node’s credential used to generate transactions (having the sender field its account address) and the other is used in the process of the block signing. Please note that this is a preliminary mechanism, in the next releases the first (private, public key) pair will be dropped when the staking mechanism will be fully implemented. To build and run the keygenerator, the following commands will need to be run:

```
$ cd ../keygenerator
$ go build
$ ./keygenerator
```

### Step 4: copying credentials and starting the node:
The previous generated .pem files needs to be copied in the same directory where the node binary resides in order to start the node.

```
$  cp initialBalancesSk.pem ./../node/config/
$  cp initialNodesSk.pem ./../node/config/
$  cd ../node && ./node
```

The node binary has some flags defined (for a brief description, the user can use --help flag). Those flags can be used to directly alter the configuration values defined in .toml/.json files and can be used when launching more than one instance of the binary. 

### Running the tests	
```	
$ go test ./...	
```

## Progress

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
  - [x] Network - Message dispatching
- [x] MetaChain
  - [x] Data Structures
  - [x] Block Processor
  - [x] Interceptors/Resolvers
  - [x] Consensus
- [x] Block K finality scheme
- [x] VM - K-Framework
  - [x] K Framework go backend
  - [x] IELE Core
  - [x] IELE Core tests
  - [x] IELE Adapter
- [x] Smart Contracts on a Sharded Architecture
  - [x] Concept reviewed
  - [x] VM integration
  - [x] SC Deployment
- [x] Governance
  - [x] Concept reviewed- 
- [x] Economics
  - [x] Concept reviewed  
- [x] Optimizations
  - [x] Randomness
  - [x] Consensus
- [x] Bootstrap from storage
- [x] Testing 
  - [x] Unit tests
  - [x] Integration tests
  - [x] TeamCity continuous integration
  - [x] Manual testing

### In progress
- [ ] Epochs
  - [ ] Nodes dispatcher (shuffling)
- [ ] Network sharding
  - [ ] Optimized Wiring protocol
- [ ] VM - K-Framework
  - [ ] EVM Core
  - [ ] EVM Core tests
  - [ ] EVM Adapter
- [ ] Smart Contracts on a Sharded Architecture
  - [ ] Dependency checker + SC migration
  - [ ] Storage rent + SC backup & restore
  - [ ] Request-response fallback
- [ ] Fee structure
- [ ] Adaptive State Sharding
  - [ ] Splitting
  - [ ] Merging 
  - [ ] Redundancy
- [ ] Privacy
- [ ] DEX integration
- [ ] Interoperability
- [ ] Optimizations
  - [ ] Smart Contract 
- [ ] Governance
  - [ ] SC for ERD IP
  - [ ] Enforced Upgrade mechanism for voted ERD IP
- [ ] Testing
  - [ ] Automate tests with AWS 
- [ ] Bugfixing
