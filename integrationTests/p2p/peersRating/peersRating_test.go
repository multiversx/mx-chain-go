package peersRating

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/factory"
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
	resolverNode := createNodeWithPeersRatingHandler(shardID, numOfShards, p2p.NormalOperation)
	maliciousNode := createNodeWithPeersRatingHandler(shardID, numOfShards, p2p.NormalOperation)
	requesterNode := createNodeWithPeersRatingHandler(core.MetachainShardId, numOfShards, p2p.NormalOperation)

	defer func() {
		_ = resolverNode.MainMessenger.Close()
		_ = maliciousNode.MainMessenger.Close()
		_ = requesterNode.MainMessenger.Close()
	}()

	time.Sleep(time.Second)
	require.Nil(t, resolverNode.ConnectOnMain(maliciousNode))
	require.Nil(t, resolverNode.ConnectOnMain(requesterNode))
	require.Nil(t, maliciousNode.ConnectOnMain(requesterNode))
	time.Sleep(time.Second)

	hdr, hdrHash, hdrBuff := getHeader()

	// Broadcasts should not be considered for peers rating
	topic := factory.ShardBlocksTopic + resolverNode.ShardCoordinator.CommunicationIdentifier(requesterNode.ShardCoordinator.SelfId())
	resolverNode.MainMessenger.Broadcast(topic, hdrBuff)
	time.Sleep(time.Second)
	maliciousNode.MainMessenger.Broadcast(topic, hdrBuff)
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

	peerRatingsMap := getRatingsMap(t, requesterNode.MainPeersRatingMonitor)
	// resolver node should have received and responded to numOfRequests
	initialResolverRating, exists := peerRatingsMap[resolverNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	initialResolverExpectedRating := fmt.Sprintf("%d", numOfRequests*(decreaseFactor+increaseFactor))
	assert.Equal(t, initialResolverExpectedRating, initialResolverRating)
	// malicious node should have only received numOfRequests
	initialMaliciousRating, exists := peerRatingsMap[maliciousNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	initialMaliciousExpectedRating := fmt.Sprintf("%d", numOfRequests*decreaseFactor)
	assert.Equal(t, initialMaliciousExpectedRating, initialMaliciousRating)

	// Reach max limits
	numOfRequests = 120
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMap(t, requesterNode.MainPeersRatingMonitor)
	// Resolver should have reached max limit and timestamps still update
	initialResolverRating, exists = peerRatingsMap[resolverNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, "100", initialResolverRating)

	// Malicious should have reached min limit and timestamps still update
	initialMaliciousRating, exists = peerRatingsMap[maliciousNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, "-100", initialMaliciousRating)

	// Add header to the malicious node's cache and remove it from the resolver's cache
	maliciousNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	resolverNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	numOfRequests = 10
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMap(t, requesterNode.MainPeersRatingMonitor)
	// resolver node should have the max rating + numOfRequests that didn't answer to
	resolverRating, exists := peerRatingsMap[resolverNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	finalResolverExpectedRating := fmt.Sprintf("%d", 100+decreaseFactor*numOfRequests)
	assert.Equal(t, finalResolverExpectedRating, resolverRating)
	// malicious node should have the min rating + numOfRequests that received and responded to
	maliciousRating, exists := peerRatingsMap[maliciousNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	finalMaliciousExpectedRating := fmt.Sprintf("%d", -100+numOfRequests*increaseFactor+(numOfRequests-1)*decreaseFactor)
	assert.Equal(t, finalMaliciousExpectedRating, maliciousRating)
}

func TestPeersRatingAndResponsivenessOnFullArchive(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	resolverFullArchiveNode := createNodeWithPeersRatingHandler(shardID, numOfShards, p2p.FullArchiveMode)
	requesterFullArchiveNode := createNodeWithPeersRatingHandler(core.MetachainShardId, numOfShards, p2p.FullArchiveMode)
	regularNode := createNodeWithPeersRatingHandler(shardID, numOfShards, p2p.FullArchiveMode)

	defer func() {
		_ = resolverFullArchiveNode.MainMessenger.Close()
		_ = resolverFullArchiveNode.FullArchiveMessenger.Close()
		_ = requesterFullArchiveNode.MainMessenger.Close()
		_ = requesterFullArchiveNode.FullArchiveMessenger.Close()
		_ = regularNode.MainMessenger.Close()
		_ = regularNode.FullArchiveMessenger.Close()
	}()

	// all nodes are connected on main network, but only the full archive resolver and requester are connected on full archive network
	time.Sleep(time.Second)
	require.Nil(t, resolverFullArchiveNode.ConnectOnFullArchive(requesterFullArchiveNode))
	require.Nil(t, resolverFullArchiveNode.ConnectOnMain(regularNode))
	require.Nil(t, requesterFullArchiveNode.ConnectOnMain(regularNode))
	time.Sleep(time.Second)

	hdr, hdrHash, hdrBuff := getHeader()

	// Broadcasts should not be considered for peers rating and should only be available on full archive network
	topic := factory.ShardBlocksTopic + resolverFullArchiveNode.ShardCoordinator.CommunicationIdentifier(requesterFullArchiveNode.ShardCoordinator.SelfId())
	resolverFullArchiveNode.FullArchiveMessenger.Broadcast(topic, hdrBuff)
	time.Sleep(time.Second)
	// check that broadcasts were successful
	_, err := requesterFullArchiveNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.Nil(t, err)
	_, err = regularNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.NotNil(t, err)
	// clean the above broadcast consequences
	requesterFullArchiveNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	resolverFullArchiveNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)

	// Broadcast on main network should also work and reach all nodes
	topic = factory.ShardBlocksTopic + regularNode.ShardCoordinator.CommunicationIdentifier(requesterFullArchiveNode.ShardCoordinator.SelfId())
	regularNode.MainMessenger.Broadcast(topic, hdrBuff)
	time.Sleep(time.Second)
	// check that broadcasts were successful
	_, err = requesterFullArchiveNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.Nil(t, err)
	_, err = resolverFullArchiveNode.DataPool.Headers().GetHeaderByHash(hdrHash)
	assert.Nil(t, err)
	// clean the above broadcast consequences
	requesterFullArchiveNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	resolverFullArchiveNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	regularNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)

	numOfRequests := 10
	// Add header to the resolver node's cache
	resolverFullArchiveNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	epochProviderStub, ok := requesterFullArchiveNode.EpochProvider.(*mock.CurrentNetworkEpochProviderStub)
	assert.True(t, ok)
	epochProviderStub.EpochIsActiveInNetworkCalled = func(epoch uint32) bool {
		return false // force the full archive requester to request from full archive network
	}
	requestHeader(requesterFullArchiveNode, numOfRequests, hdrHash, resolverFullArchiveNode.ShardCoordinator.SelfId())

	peerRatingsMap := getRatingsMap(t, requesterFullArchiveNode.FullArchivePeersRatingMonitor)
	// resolver node should have received and responded to numOfRequests
	initialResolverRating, exists := peerRatingsMap[resolverFullArchiveNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	initialResolverExpectedRating := fmt.Sprintf("%d", numOfRequests*(decreaseFactor+increaseFactor))
	assert.Equal(t, initialResolverExpectedRating, initialResolverRating)
	// main nodes should not be found in this cacher
	_, exists = peerRatingsMap[regularNode.MainMessenger.ID().Pretty()]
	require.False(t, exists)

	// force the full archive requester to request the header from main network
	// as it does not exists on the main resolver, it should only decrease its rating
	epochProviderStub.EpochIsActiveInNetworkCalled = func(epoch uint32) bool {
		return true // force the full archive requester to request from main network
	}
	requestHeader(requesterFullArchiveNode, numOfRequests, hdrHash, regularNode.ShardCoordinator.SelfId())
	peerRatingsMap = getRatingsMap(t, requesterFullArchiveNode.MainPeersRatingMonitor)

	_, exists = peerRatingsMap[resolverFullArchiveNode.MainMessenger.ID().Pretty()]
	require.False(t, exists) // should not be any request on the main monitor to the full archive resolver

	mainResolverRating, exists := peerRatingsMap[regularNode.MainMessenger.ID().Pretty()]
	require.True(t, exists)
	mainResolverExpectedRating := fmt.Sprintf("%d", numOfRequests*decreaseFactor)
	assert.Equal(t, mainResolverExpectedRating, mainResolverRating)
}

func createNodeWithPeersRatingHandler(shardID uint32, numShards uint32, nodeOperationMode p2p.NodeOperation) *integrationTests.TestProcessorNode {

	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:              numShards,
		NodeShardId:            shardID,
		WithPeersRatingHandler: true,
		NodeOperationMode:      nodeOperationMode,
	})

	return tpn
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

func getRatingsMap(t *testing.T, monitor p2p.PeersRatingMonitor) map[string]string {
	peerRatingsStr := monitor.GetConnectedPeersRatings()
	peerRatingsMap := make(map[string]string)

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
