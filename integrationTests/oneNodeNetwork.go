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

	network := &oneNodeNetwork{
		advertiser:   CreateMessengerWithKadDht(""),
		IdxProposers: make([]int, numOfShards+1),
	}

	network.Nodes = CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		GetConnectableAddress(network.advertiser),
	)

	for i := 0; i < numOfShards; i++ {
		network.IdxProposers[i] = i * nodesPerShard
	}

	network.Node = network.Nodes[0]
	network.IdxProposers[numOfShards] = numOfShards * nodesPerShard

	return network
}

func (network *oneNodeNetwork) Start() {
	_ = network.advertiser.Bootstrap()
	DisplayAndStartNodes(network.Nodes)
}

func (network *oneNodeNetwork) Stop() {
	defer func() {
		_ = network.advertiser.Close()
		for _, n := range network.Nodes {
			_ = n.Messenger.Close()
		}
	}()
}

func (network *oneNodeNetwork) Mint(address []byte, value *big.Int) {
	MintAddress(network.Node.AccntState, address, value)
}

func (network *oneNodeNetwork) GetMinGasPrice() uint64 {
	return network.Node.EconomicsData.GetMinGasPrice()
}

func (network *oneNodeNetwork) MaxGasLimitPerBlock() uint64 {
	return network.Node.EconomicsData.MaxGasLimitPerBlock(0) - 1
}

func (network *oneNodeNetwork) GoToRoundOne() {
	network.Round = IncrementAndPrintRound(network.Round)
	network.Nonce++
}

func (network *oneNodeNetwork) Continue(t *testing.T, numRounds int) {
	network.Nonce, network.Round = WaitOperationToBeDone(t, network.Nodes, numRounds, network.Nonce, network.Round, network.IdxProposers)
}

func (network *oneNodeNetwork) AddTxToPool(tx *transaction.Transaction) {
	txHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, tx)
	sourceShard := network.Node.ShardCoordinator.ComputeId(tx.SndAddr)
	cacheIdentifier := process.ShardCacherIdentifier(sourceShard, sourceShard)
	network.Node.DataPool.Transactions().AddData(txHash, tx, tx.Size(), cacheIdentifier)
}
