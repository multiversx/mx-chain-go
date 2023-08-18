package request

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestsInShardingEnvironmentWithConnectNewNode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	) // all nodes are connected in a complete graph manner

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	mb := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("tx hash1")},
		ReceiverShardID: 1,
		SenderShardID:   0,
		Type:            0,
		Reserved:        nil,
	}
	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, mb)
	require.Nil(t, err)

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != 0 {
			continue
		}

		// shard 0 nodes will contain this miniblock
		node.DataPool.MiniBlocks().Put(hash, mb, 0)
	}

	// the shard nodes will request this header
	var lastIdxMissing int
	for i := 0; i < 10; i++ {
		requestHashOnShardNodes(nodes, hash)
		lastIdxMissing = checkAllNodesHaveBlock(nodes, hash)
		if lastIdxMissing == -1 {
			break
		}

		time.Sleep(time.Second)
	}
	assert.Equal(t, -1, lastIdxMissing)

	args := integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numOfShards),
		NodeShardId:          1,
		TxSignPrivKeyShardId: 0,
		EpochsConfig:         integrationTests.GetDefaultEnableEpochsConfig(),
	}
	newNode := integrationTests.NewTestProcessorNode(args)
	defer newNode.Close()

	allNodes := []integrationTests.Connectable{newNode}
	for _, n := range nodes {
		allNodes = append(allNodes, n)
	}
	integrationTests.ConnectNodes(allNodes)

	onlyNewNodeList := []*integrationTests.TestProcessorNode{newNode}
	integrationTests.DisplayAndStartNodes(onlyNewNodeList)

	// the shard nodes will request this miniblock
	for i := 0; i < 10; i++ {
		requestHashOnShardNodes(onlyNewNodeList, hash)
		lastIdxMissing = checkAllNodesHaveBlock(onlyNewNodeList, hash)
		if lastIdxMissing == -1 {
			break
		}

		time.Sleep(time.Second)
	}
	assert.Equal(t, -1, lastIdxMissing)
}

func requestHashOnShardNodes(nodes []*integrationTests.TestProcessorNode, hash []byte) {
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != 1 {
			continue
		}

		node.RequestHandler.RequestMiniBlock(0, hash)
	}
}

func checkAllNodesHaveBlock(nodes []*integrationTests.TestProcessorNode, hash []byte) int {
	for idx, n := range nodes {
		if (n.ShardCoordinator.SelfId() != 0) && (n.ShardCoordinator.SelfId() != 1) {
			continue
		}

		_, ok := n.DataPool.MiniBlocks().Get(hash)
		if !ok {
			return idx
		}
	}

	return -1
}
