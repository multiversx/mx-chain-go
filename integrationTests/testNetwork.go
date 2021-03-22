package integrationTests

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/require"
)

type testNetwork struct {
	NumShards         int
	NodesPerShard     int
	NodesInMetashard  int
	Nodes             []*TestProcessorNode
	Proposers         []int
	Advertiser        p2p.Messenger
	AdvertiserAddress string
	Round             uint64
	Nonce             uint64
	T                 *testing.T
}

func NewTestNetwork(t *testing.T) *testNetwork {
	// TODO replace testing.T with testing.TB everywhere in the integrationTest utilities
	return &testNetwork{
		T: t,
	}
}

func (net *testNetwork) Start() {
	net.createAdvertiser()
	net.createNodes()
	net.indexProposers()
	net.startNodes()

	net.Round = 0
	net.Nonce = 0
}

func (net *testNetwork) Step() {
	net.Round = IncrementAndPrintRound(net.Round)
	net.Nonce++
}

func (net *testNetwork) Steps(steps int) {
	net.Nonce, net.Round = WaitOperationToBeDone(
		net.T,
		net.Nodes,
		steps,
		net.Nonce,
		net.Round,
		net.Proposers,
	)
}

func (net *testNetwork) Close() {
	net.closeAdvertiser()
	net.closeNodes()
}

func (net *testNetwork) MintNodeAccounts(value *big.Int) {
	MintAllNodes(net.Nodes, value)
}

func (net *testNetwork) MintNodeAccountsInt64(value int64) {
	MintAllNodes(net.Nodes, big.NewInt(value))
}

func (net *testNetwork) createAdvertiser() {
	net.Advertiser = CreateMessengerWithKadDht(net.AdvertiserAddress)
	err := net.Advertiser.Bootstrap()
	require.Nil(net.T, err)
}

func (net *testNetwork) createNodes() {
	net.Nodes = CreateNodes(
		net.NumShards,
		net.NodesPerShard,
		net.NodesInMetashard,
		GetConnectableAddress(net.Advertiser),
	)
}

func (net *testNetwork) indexProposers() {
	net.Proposers = make([]int, net.NumShards+1)
	for i := 0; i < net.NumShards; i++ {
		net.Proposers[i] = i * net.NumShards
	}
	net.Proposers[net.NumShards] = net.NumShards * net.NodesPerShard
}

func (net *testNetwork) startNodes() {
	DisplayAndStartNodes(net.Nodes)
}

func (net *testNetwork) closeAdvertiser() {
	err := net.Advertiser.Close()
	require.Nil(net.T, err)
}

func (net *testNetwork) closeNodes() {
	for _, node := range net.Nodes {
		err := node.Messenger.Close()
		require.Nil(net.T, err)
	}
}
