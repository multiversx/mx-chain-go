package integrationTests

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type oneNodeNetwork struct {
	Round uint64
	Nonce uint64

	Node *TestProcessorNode
}

// NewOneNodeNetwork creates a one-node network, useful for some integration tests
func NewOneNodeNetwork() *oneNodeNetwork {
	n := &oneNodeNetwork{}

	nodes := CreateNodes(
		1,
		1,
		0,
		"address",
	)

	n.Node = nodes[0]
	return n
}

// Stop stops the test network
func (n *oneNodeNetwork) Stop() {
	_ = n.Node.Messenger.Close()
}

// Mint mints the given address
func (n *oneNodeNetwork) Mint(address []byte, value *big.Int) {
	MintAddress(n.Node.AccntState, address, value)
}

// GetMinGasPrice returns the min gas price
func (n *oneNodeNetwork) GetMinGasPrice() uint64 {
	return n.Node.EconomicsData.GetMinGasPrice()
}

// MaxGasLimitPerBlock returns the max gas per block
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
	n.Nonce, n.Round = WaitOperationToBeDone(t, []*TestProcessorNode{n.Node}, numRounds, n.Nonce, n.Round, []int{0})
}

// AddTxToPool adds a transaction to the pool (skips signature checks and interceptors)
func (n *oneNodeNetwork) AddTxToPool(tx *transaction.Transaction) {
	txHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, tx)
	sourceShard := n.Node.ShardCoordinator.ComputeId(tx.SndAddr)
	cacheIdentifier := process.ShardCacherIdentifier(sourceShard, sourceShard)
	n.Node.DataPool.Transactions().AddData(txHash, tx, tx.Size(), cacheIdentifier)
}
