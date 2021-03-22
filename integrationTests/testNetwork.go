package integrationTests

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/require"
)

type TestNetwork struct {
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

func NewTestNetwork(t *testing.T) *TestNetwork {
	// TODO replace testing.T with testing.TB everywhere in the integrationTest utilities
	return &TestNetwork{
		T: t,
	}
}

func (net *TestNetwork) Start() {
	net.createAdvertiser()
	net.createNodes()
	net.indexProposers()
	net.startNodes()

	net.Round = 0
	net.Nonce = 0
}

func (net *TestNetwork) Step() {
	net.Round = IncrementAndPrintRound(net.Round)
	net.Nonce++
}

func (net *TestNetwork) Steps(steps int) {
	net.Nonce, net.Round = WaitOperationToBeDone(
		net.T,
		net.Nodes,
		steps,
		net.Nonce,
		net.Round,
		net.Proposers,
	)
}

func (net *TestNetwork) Close() {
	net.closeAdvertiser()
	net.closeNodes()
}

func (net *TestNetwork) MintNodeAccounts(value *big.Int) {
	MintAllNodes(net.Nodes, value)
}

func (net *TestNetwork) MintNodeAccountsInt64(value int64) {
	MintAllNodes(net.Nodes, big.NewInt(value))
}

func (net *TestNetwork) createAdvertiser() {
	net.Advertiser = CreateMessengerWithKadDht(net.AdvertiserAddress)
	err := net.Advertiser.Bootstrap(0)
	require.Nil(net.T, err)
}

func (net *TestNetwork) createNodes() {
	net.Nodes = CreateNodes(
		net.NumShards,
		net.NodesPerShard,
		net.NodesInMetashard,
		GetConnectableAddress(net.Advertiser),
	)
}

func (net *TestNetwork) indexProposers() {
	net.Proposers = make([]int, net.NumShards+1)
	for i := 0; i < net.NumShards; i++ {
		net.Proposers[i] = i * net.NumShards
	}
	net.Proposers[net.NumShards] = net.NumShards * net.NodesPerShard
}

func (net *TestNetwork) startNodes() {
	DisplayAndStartNodes(net.Nodes)
}

func (net *TestNetwork) closeAdvertiser() {
	err := net.Advertiser.Close()
	require.Nil(net.T, err)
}

func (net *TestNetwork) closeNodes() {
	for _, node := range net.Nodes {
		err := node.Messenger.Close()
		require.Nil(net.T, err)
	}
}
