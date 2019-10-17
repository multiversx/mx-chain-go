package block

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

const broadcastDelay = 2 * time.Second

func TestInterceptedShardBlockHeaderVerifiedWithCorrectConsensusGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, randomness)

	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		v, ok := metaNode.MetaDataPool.ShardHeaders().Get(headerHash)
		assert.True(t, ok)
		assert.Equal(t, header, v)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		v, ok := shardNode.ShardDataPool.Headers().Get(headerHash)
		assert.True(t, ok)
		assert.Equal(t, header, v)
	}
}

func TestInterceptedMetaBlockVerifiedWithCorrectConsensusGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Metachain node Generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(
		sharding.MetachainShardId,
		nodesMap,
		round,
		nonce,
		randomness,
	)

	nodesMap[sharding.MetachainShardId][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		v, ok := metaNode.MetaDataPool.MetaBlocks().Get(headerHash)
		assert.True(t, ok)
		assert.Equal(t, header, v)
	}

	// all nodes in shard do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, shardNode := range nodesMap[0] {
		v, ok := shardNode.ShardDataPool.MetaBlocks().Get(headerHash)
		assert.True(t, ok)
		assert.Equal(t, header, v)
	}
}
