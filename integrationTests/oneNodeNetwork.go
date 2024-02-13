package integrationTests

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
)

// OneNodeNetwork is a one-node network, useful for some integration tests
type OneNodeNetwork struct {
	Round uint64
	Nonce uint64

	Node *TestProcessorNode
}

// NewOneNodeNetwork creates a OneNodeNetwork
func NewOneNodeNetwork() *OneNodeNetwork {
	n := &OneNodeNetwork{}

	nodes := CreateNodes(
		1,
		1,
		0,
	)

	n.Node = nodes[0]
	return n
}

// Stop stops the test network
func (n *OneNodeNetwork) Stop() {
	n.Node.Close()
}

// Mint mints the given address
func (n *OneNodeNetwork) Mint(address []byte, value *big.Int) {
	MintAddress(n.Node.AccntState, address, value)
}

// GetMinGasPrice returns the min gas price
func (n *OneNodeNetwork) GetMinGasPrice() uint64 {
	return n.Node.EconomicsData.GetMinGasPrice()
}

// MaxGasLimitPerBlock returns the max gas per block
func (n *OneNodeNetwork) MaxGasLimitPerBlock() uint64 {
	return n.Node.EconomicsData.MaxGasLimitPerBlock(0) - 1
}

// GoToRoundOne advances processing to block and round 1
func (n *OneNodeNetwork) GoToRoundOne() {
	n.Round = IncrementAndPrintRound(n.Round)
	n.Nonce++
}

// Continue advances processing with a number of rounds
func (n *OneNodeNetwork) Continue(t *testing.T, numRounds int) {
	n.Nonce, n.Round = WaitOperationToBeDone(t, []*TestProcessorNode{n.Node}, numRounds, n.Nonce, n.Round, []int{0})
}

// AddTxToPool adds a transaction to the pool (skips signature checks and interceptors)
func (n *OneNodeNetwork) AddTxToPool(tx *transaction.Transaction) {
	txHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, tx)
	sourceShard := n.Node.ShardCoordinator.ComputeId(tx.SndAddr)
	cacheIdentifier := process.ShardCacherIdentifier(sourceShard, sourceShard)
	n.Node.DataPool.Transactions().AddData(txHash, tx, tx.Size(), cacheIdentifier)
}
