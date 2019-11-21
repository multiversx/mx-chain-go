package block

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
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

func TestInterceptedShardBlockHeaderWithLeaderSignatureAndRandSeedChecks(t *testing.T) {
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

	singleSigner := getBlockSingleSignerStub()
	keyGen := signing.NewKeyGenerator(kyber.NewSuitePairingBn256())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
		singleSigner,
		keyGen,
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

func TestInterceptedShardHeaderBlockWithWrongPreviousRandSeendShouldNotBeAccepted(t *testing.T) {
	nodesPerShard := 4
	nbMetaNodes := 4
	nbShards := 1
	consensusGroupSize := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	singleSigner := getBlockSingleSignerStub()
	keyGen := signing.NewKeyGenerator(kyber.NewSuitePairingBn256())
	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
		singleSigner,
		keyGen,
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

	wrongRandomness := []byte("wrong randomness")
	round := uint64(2)
	nonce := uint64(2)
	body, header, _, _ := integrationTests.ProposeBlockWithConsensusSignature(0, nodesMap, round, nonce, wrongRandomness)

	nodesMap[0][0].BroadcastBlock(body, header)

	time.Sleep(broadcastDelay)

	headerBytes, _ := integrationTests.TestMarshalizer.Marshal(header)
	headerHash := integrationTests.TestHasher.Compute(string(headerBytes))

	// all nodes in metachain have the block header in pool as interceptor validates it
	for _, metaNode := range nodesMap[sharding.MetachainShardId] {
		_, ok := metaNode.MetaDataPool.ShardHeaders().Get(headerHash)
		assert.False(t, ok)
	}

	// all nodes in shard have the block in pool as interceptor validates it
	for _, shardNode := range nodesMap[0] {
		_, ok := shardNode.ShardDataPool.Headers().Get(headerHash)
		assert.False(t, ok)
	}
}

func getBlockSingleSignerStub() crypto.SingleSigner {
	singleSigner := singlesig.BlsSingleSigner{}
	return &mock.SignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return singleSigner.Sign(private, msg)
		},
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			// if rand seed is verified, return nil if the message contains "random seed"
			if strings.Contains(string(msg), "random seed") {
				return nil
			}
			// if leader signature is verified, return nil if the signature contains "leader sign"
			if strings.Contains(string(sig), "leader sign") {
				var hdr block.Header
				// if the marshalized struct is a header, everything is ok
				err := json.Unmarshal(msg, &hdr)
				return err
			}
			return singleSigner.Verify(public, msg, sig)
		},
	}
}
