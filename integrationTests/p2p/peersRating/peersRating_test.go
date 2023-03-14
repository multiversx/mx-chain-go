package peersRating

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	increaseFactor = 2
	decreaseFactor = -1
)

func TestPeersRatingAndResponsiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	resolverNode := createNodeWithPeersRatingHandler(shardID, numOfShards)
	maliciousNode := createNodeWithPeersRatingHandler(shardID, numOfShards)
	requesterNode := createNodeWithPeersRatingHandler(core.MetachainShardId, numOfShards)

	defer func() {
		_ = resolverNode.Messenger.Close()
		_ = maliciousNode.Messenger.Close()
		_ = requesterNode.Messenger.Close()
	}()

	time.Sleep(time.Second)
	require.Nil(t, resolverNode.ConnectTo(maliciousNode))
	require.Nil(t, resolverNode.ConnectTo(requesterNode))
	require.Nil(t, maliciousNode.ConnectTo(requesterNode))
	time.Sleep(time.Second)

	hdr, hdrHash, hdrBuff := getHeader()

	// Broadcasts should not be considered for peers rating
	topic := factory.ShardBlocksTopic + resolverNode.ShardCoordinator.CommunicationIdentifier(requesterNode.ShardCoordinator.SelfId())
	resolverNode.Messenger.Broadcast(topic, hdrBuff)
	time.Sleep(time.Second)
	maliciousNode.Messenger.Broadcast(topic, hdrBuff)
	time.Sleep(time.Second)
	// check that broadcasts were successful
	_, err := requesterNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.Nil(t, err)
	_, err = maliciousNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.Nil(t, err)
	// clean the above broadcasts consequences as only resolverNode should have the header
	requesterNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	maliciousNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)

	numOfRequests := 10
	// Add header to the resolver node's cache
	resolverNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap := getRatingsMap(t, requesterNode)
	// resolver node should have received and responded to numOfRequests
	initialResolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialResolverExpectedRating := numOfRequests * (decreaseFactor + increaseFactor)
	assert.Equal(t, int32(initialResolverExpectedRating), initialResolverRating)
	// malicious node should have only received numOfRequests
	initialMaliciousRating, exists := peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialMaliciousExpectedRating := numOfRequests * decreaseFactor
	assert.Equal(t, int32(initialMaliciousExpectedRating), initialMaliciousRating)

	// Reach max limits
	numOfRequests = 120
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMap(t, requesterNode)
	// Resolver should have reached max limit and timestamps still update
	initialResolverRating, exists = peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, int32(100), initialResolverRating)

	// Malicious should have reached min limit and timestamps still update
	initialMaliciousRating, exists = peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, int32(-100), initialMaliciousRating)

	// Add header to the malicious node's cache and remove it from the resolver's cache
	maliciousNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	resolverNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	numOfRequests = 10
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMap(t, requesterNode)
	// resolver node should have the max rating + numOfRequests that didn't answer to
	resolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	finalResolverExpectedRating := 100 + decreaseFactor*numOfRequests
	assert.Equal(t, int32(finalResolverExpectedRating), resolverRating)
	// malicious node should have the min rating + numOfRequests that received and responded to
	maliciousRating, exists := peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	finalMaliciousExpectedRating := -100 + numOfRequests*increaseFactor + (numOfRequests-1)*decreaseFactor
	assert.Equal(t, int32(finalMaliciousExpectedRating), maliciousRating)
}

func createNodeWithPeersRatingHandler(shardID uint32, numShards uint32) *integrationTests.TestProcessorNode {
	topRatedCache := testscommon.NewCacherMock()
	badRatedCache := testscommon.NewCacherMock()
	peersRatingHandler, _ := p2pFactory.NewPeersRatingHandler(
		p2pFactory.ArgPeersRatingHandler{
			TopRatedCache: topRatedCache,
			BadRatedCache: badRatedCache,
		})

	peersRatingMonitor, _ := p2pFactory.NewPeersRatingMonitor(
		p2pFactory.ArgPeersRatingMonitor{
			TopRatedCache: topRatedCache,
			BadRatedCache: badRatedCache,
		})

	return integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:          numShards,
		NodeShardId:        shardID,
		PeersRatingHandler: peersRatingHandler,
		PeersRatingMonitor: peersRatingMonitor,
	})
}

func getHeader() (*block.Header, []byte, []byte) {
	hdr := &block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardID:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 1),
		RandSeed:         make([]byte, 1),
		MiniBlockHeaders: nil,
		ChainID:          integrationTests.ChainID,
		SoftwareVersion:  []byte("version"),
		AccumulatedFees:  big.NewInt(100),
		DeveloperFees:    big.NewInt(10),
	}
	hdrBuff, _ := integrationTests.TestMarshalizer.Marshal(hdr)
	hdrHash := integrationTests.TestHasher.Compute(string(hdrBuff))
	return hdr, hdrHash, hdrBuff
}

func getRatingsMap(t *testing.T, node *integrationTests.TestProcessorNode) map[string]int32 {
	peerRatingsStr := node.PeersRatingMonitor.GetPeersRatings()
	peerRatingsMap := make(map[string]int32)

	err := json.Unmarshal([]byte(peerRatingsStr), &peerRatingsMap)
	require.Nil(t, err)

	return peerRatingsMap
}

func requestHeader(requesterNode *integrationTests.TestProcessorNode, numOfRequests int, hdrHash []byte, shardID uint32) {
	for i := 0; i < numOfRequests; i++ {
		requesterNode.RequestHandler.RequestShardHeader(shardID, hdrHash)
		time.Sleep(time.Second) // allow nodes to respond
	}
}
