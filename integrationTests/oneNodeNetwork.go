package integrationTests

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

type oneNodeNetwork struct {
	advertiser   p2p.Messenger
	Nodes        []*TestProcessorNode
	IdxProposers []int
	Round        uint64
	Nonce        uint64

	Node *TestProcessorNode
}

// NewOneNodeNetwork creates a one-node network, useful for some integration tests
func NewOneNodeNetwork() *oneNodeNetwork {
	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 0

	n := &oneNodeNetwork{
		advertiser:   CreateMessengerWithKadDht(""),
		IdxProposers: make([]int, numOfShards+1),
	}

	n.Nodes = CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		GetConnectableAddress(n.advertiser),
	)

	for i := 0; i < numOfShards; i++ {
		n.IdxProposers[i] = i * nodesPerShard
	}

	n.Node = n.Nodes[0]
	n.IdxProposers[numOfShards] = numOfShards * nodesPerShard

	return n
}

// Start starts the test network
func (n *oneNodeNetwork) Start() {
	_ = n.advertiser.Bootstrap()
	DisplayAndStartNodes(n.Nodes)
}

// Stop stops the test network
func (n *oneNodeNetwork) Stop() {
	defer func() {
		_ = n.advertiser.Close()
		for _, n := range n.Nodes {
			_ = n.Messenger.Close()
		}
	}()
}

// Mint mints the given address
func (n *oneNodeNetwork) Mint(address []byte, value *big.Int) {
	MintAddress(n.Node.AccntState, address, value)
}

// GetMinGasPrice returns the min gas price
func (n *oneNodeNetwork) GetMinGasPrice() uint64 {
	return n.Node.EconomicsData.GetMinGasPrice()
}

// GetMinGasPrice returns the max gas per block
func (n *oneNodeNetwork) MaxGasLimitPerBlock() uint64 {
	return n.Node.EconomicsData.MaxGasLimitPerBlock(0) - 1
}

// GoToRoundOne advances processing to block and round 1
func (n *oneNodeNetwork) GoToRoundOne() {
	n.Round = IncrementAndPrintRound(n.Round)
	n.Nonce++
}

// Continue advances processing with a number of rounds
func (n *oneNodeNetwork) Continue(t *testing.T, numRounds int) {
	n.Nonce, n.Round = WaitOperationToBeDone(t, n.Nodes, numRounds, n.Nonce, n.Round, n.IdxProposers)
}

// AddTxToPool adds a transaction to the pool (skips signature checks and interceptors)
func (n *oneNodeNetwork) AddTxToPool(tx *transaction.Transaction) {
	txHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, tx)
	sourceShard := n.Node.ShardCoordinator.ComputeId(tx.SndAddr)
	cacheIdentifier := process.ShardCacherIdentifier(sourceShard, sourceShard)
	n.Node.DataPool.Transactions().AddData(txHash, tx, tx.Size(), cacheIdentifier)
}
