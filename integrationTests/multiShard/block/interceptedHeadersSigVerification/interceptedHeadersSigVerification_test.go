package interceptedHeadersSigVerification

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclsinglesig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
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
	singleSigner := &mclsinglesig.BlsSingleSigner{}

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	var err error
	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, randomness, 0)
	header, err = fillHeaderFields(nodesMap[0][0], header, singleSigner)
	assert.Nil(t, err)

	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[core.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
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

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Metachain node Generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(
		core.MetachainShardId,
		nodesMap,
		round,
		nonce,
		randomness,
		0,
	)

	nodesMap[core.MetachainShardId][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))
	hmb := header.(*block.MetaBlock)

	// all nodes in metachain do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, metaNode := range nodesMap[core.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.True(t, hmb.Equal(v))
	}

	// all nodes in shard do not have the block in pool as interceptor does not validate it with a wrong consensus
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.True(t, hmb.Equal(v))
	}
}

func TestInterceptedShardBlockHeaderWithLeaderSignatureAndRandSeedChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	singleSigner := &mclsinglesig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		singleSigner,
		keyGen,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	// one testNodeProcessor from shard proposes block signed by all other nodes in shard consensus
	randomness := []byte("random seed")
	round := uint64(1)
	nonce := uint64(1)

	body, header, _, consensusNodes := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, randomness, 0)
	nodeToSendFrom := consensusNodes[0]
	err := header.SetPrevRandSeed(randomness)
	assert.Nil(t, err)

	header, err = fillHeaderFields(nodeToSendFrom, header, singleSigner)
	assert.Nil(t, err)

	nodeToSendFrom.BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[core.MetachainShardId] {
		v, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		v, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Nil(t, err)
		assert.Equal(t, header, v)
	}
}

func TestInterceptedShardHeaderBlockWithWrongPreviousRandSeedShouldNotBeAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	singleSigner := &mclsinglesig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		singleSigner,
		keyGen,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Shard node generating header and block body...")

	wrongRandomness := []byte("wrong randomness")
	round := uint64(2)
	nonce := uint64(2)
	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, wrongRandomness, 0)

	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[core.MetachainShardId] {
		_, err := metaNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Error(t, err)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		_, err := shardNode.DataPool.Headers().GetHeaderByHash(headerHash)
		assert.Error(t, err)
	}
}

func fillHeaderFields(proposer *integrationTests.TestProcessorNode, hdr data.HeaderHandler, signer crypto.SingleSigner) (data.HeaderHandler, error) {
	leaderSk := proposer.NodeKeys.Sk

	randSeed, _ := signer.Sign(leaderSk, hdr.GetPrevRandSeed())
	err := hdr.SetRandSeed(randSeed)
	if err != nil {
		return nil, err
	}

	hdrClone := hdr.ShallowClone()
	err = hdrClone.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
	}

	headerJsonBytes, _ := integrationTests.TestMarshalizer.Marshal(hdrClone)
	leaderSign, _ := signer.Sign(leaderSk, headerJsonBytes)
	err = hdr.SetLeaderSignature(leaderSign)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}
