<div style="text-align:center">
  <img src="https://github.com/ElrondNetwork/elrond-go/blob/master/elrond_logo_01.svg"></img>
</div>  

<br>

[![](https://img.shields.io/badge/made%20by-Elrond%20Network-blue.svg)](http://elrond.com/)
[![](https://img.shields.io/badge/project-Elrond%20Network%20Testnet-blue.svg)](http://elrond.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ElrondNetwork/elrond-go)](https://goreportcard.com/report/github.com/ElrondNetwork/elrond-go)
[![LoC](https://tokei.rs/b1/github/ElrondNetwork/elrond-go?category=code)](https://github.com/ElrondNetwork/elrond-go)
[![API Reference](https://godoc.org/github.com/ElrondNetwork/elrond-go?status.svg)](https://godoc.org/github.com/ElrondNetwork/elrond-go)
[![riot.im](https://img.shields.io/badge/riot.im-JOIN%20CHAT-green.svg)](https://riot.im/app/#/room/#elrond:matrix.org)

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
$ git clone https://github.com/ElrondNetwork/elrond-go
$ cd elrond-go && git checkout master
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

### Start the node 
#### Step 4a: Join Elrond testnet:
Follow the steps outlined [here](https://docs.elrond.com/start-a-validator-node/start-the-network). This is because in order to join the testnet you need a specific node configuration.
______
OR
______
#### Step 4b: copying credentials and starting a node in a separate network:
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
  - [x] Concept reviewed
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
  - [ ] Optimized wiring protocol
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


## Contribution
Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes to Elrond!

If you'd like to contribute to Elrond, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core developers first on our [riot channel](https://riot.im/app/#/room/#elrond:matrix.org) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 - Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines.
 - Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 - Pull requests need to be based on and opened against the master branch.
 - Commit messages should be prefixed with the package(s) they modify.
    - E.g. "core/indexer: fixed a typo"

Please see the [documentation](https://docs.elrond.com/) for more details on the Elrond project.
