package request

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestsInShardingEnvironment(t *testing.T) {
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

	// metachain nodes will contain a dummy metablock in their caches
	metablock := &block.MetaBlock{
		Nonce:                  1,
		Round:                  1,
		TimeStamp:              uint64(time.Now().Unix()),
		RootHash:               []byte("root hash"),
		PubKeysBitmap:          []byte{0},
		PrevHash:               []byte("prev hash"),
		Signature:              []byte("signature"),
		RandSeed:               []byte("rand seed"),
		PrevRandSeed:           []byte("prev rand seed"),
		ValidatorStatsRootHash: []byte("validator root hash"),
		ChainID:                integrationTests.ChainID,
		SoftwareVersion:        []byte("version"),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
	}
	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metablock)
	require.Nil(t, err)

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		node.DataPool.Headers().AddHeader(hash, metablock)
	}

	// the shard nodes will request this metablock
	var lastIdxMissing int
	var ok bool
	for i := 0; i < 10; i++ {
		requestHashOnShardNodes(nodes, hash)
		lastIdxMissing, ok = checkAllNodesHaveBlock(nodes, hash)
		if ok {
			return
		}

		time.Sleep(time.Second)
	}

	assert.Fail(t, fmt.Sprintf("node iwth index %d did not received the metablock hash %x", lastIdxMissing, hash))
}

func requestHashOnShardNodes(nodes []*integrationTests.TestProcessorNode, hash []byte) {
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == core.MetachainShardId {
			continue
		}

		node.RequestHandler.RequestMetaHeader(hash)
	}
}

func checkAllNodesHaveBlock(nodes []*integrationTests.TestProcessorNode, hash []byte) (int, bool) {
	for idx, n := range nodes {
		_, err := n.DataPool.Headers().GetHeaderByHash(hash)
		if err != nil {
			return idx, false
		}
	}

	return 0, true
}
