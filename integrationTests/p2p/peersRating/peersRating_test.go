package peersRating

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	p2pFactory "github.com/ElrondNetwork/elrond-go/p2p/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	increaseFactor = 2
	decreaseFactor = -1
)

type ratingInfo struct {
	Rating                       int32 `json:"rating"`
	TimestampLastRequestToPid    int64 `json:"timestampLastRequestToPid"`
	TimestampLastResponseFromPid int64 `json:"timestampLastResponseFromPid"`
}

func TestPeersRatingAndResponsiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	peersRatingCfg := config.PeersRatingConfig{
		TimeWaitingForReconnectionInSec: 300, // not looking to clean cachers on this test
		TimeBetweenMetricsUpdateInSec:   1,
		TimeBetweenCachersSweepInSec:    15,
	}
	resolverNode := createNodeWithPeersRatingHandler(shardID, numOfShards, peersRatingCfg)
	maliciousNode := createNodeWithPeersRatingHandler(shardID, numOfShards, peersRatingCfg)
	requesterNode := createNodeWithPeersRatingHandler(core.MetachainShardId, numOfShards, peersRatingCfg)

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

	hdr, hdrHash := getHeader()

	numOfRequests := 10
	// Add header to the resolver node's cache
	resolverNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap := getRatingsMapFromMetric(t, requesterNode)
	// resolver node should have received and responded to numOfRequests
	initialResolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialResolverExpectedRating := numOfRequests * (decreaseFactor + increaseFactor)
	assert.Equal(t, int32(initialResolverExpectedRating), initialResolverRating.Rating)
	testTimestampsForRespondingNode(t, initialResolverRating.TimestampLastResponseFromPid, initialResolverRating.TimestampLastRequestToPid)
	// malicious node should have only received numOfRequests
	initialMaliciousRating, exists := peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialMaliciousExpectedRating := numOfRequests * decreaseFactor
	assert.Equal(t, int32(initialMaliciousExpectedRating), initialMaliciousRating.Rating)
	testTimestampsForNotRespondingNode(t, initialMaliciousRating.TimestampLastResponseFromPid, initialMaliciousRating.TimestampLastRequestToPid, 0)

	// Reach max limits
	numOfRequests = 120
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMapFromMetric(t, requesterNode)
	// Resolver should have reached max limit and timestamps still update
	initialResolverRating, exists = peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, int32(100), initialResolverRating.Rating)
	testTimestampsForRespondingNode(t, initialResolverRating.TimestampLastResponseFromPid, initialResolverRating.TimestampLastRequestToPid)

	// Malicious should have reached min limit and timestamps still update
	initialMaliciousRating, exists = peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, int32(-100), initialMaliciousRating.Rating)
	testTimestampsForNotRespondingNode(t, initialMaliciousRating.TimestampLastResponseFromPid, initialMaliciousRating.TimestampLastRequestToPid, 0)

	// Add header to the malicious node's cache and remove it from the resolver's cache
	maliciousNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	resolverNode.DataPool.Headers().RemoveHeaderByHash(hdrHash)
	numOfRequests = 10
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap = getRatingsMapFromMetric(t, requesterNode)
	// resolver node should have the max rating + numOfRequests that didn't answer to
	resolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	finalResolverExpectedRating := 100 + decreaseFactor*numOfRequests
	assert.Equal(t, int32(finalResolverExpectedRating), resolverRating.Rating)
	testTimestampsForNotRespondingNode(t, resolverRating.TimestampLastResponseFromPid, resolverRating.TimestampLastRequestToPid, initialResolverRating.TimestampLastResponseFromPid)
	// malicious node should have the min rating + numOfRequests that received and responded to
	maliciousRating, exists := peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	finalMaliciousExpectedRating := -100 + numOfRequests*increaseFactor + (numOfRequests-1)*decreaseFactor
	assert.Equal(t, int32(finalMaliciousExpectedRating), maliciousRating.Rating)
	testTimestampsForRespondingNode(t, maliciousRating.TimestampLastResponseFromPid, maliciousRating.TimestampLastRequestToPid)
}

func TestPeersRatingAndCachersCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	peersRatingCfg := config.PeersRatingConfig{
		TimeWaitingForReconnectionInSec: 12,
		TimeBetweenMetricsUpdateInSec:   1,
		TimeBetweenCachersSweepInSec:    2,
	}
	resolverNode := createNodeWithPeersRatingHandler(shardID, numOfShards, peersRatingCfg)
	maliciousNode := createNodeWithPeersRatingHandler(shardID, numOfShards, peersRatingCfg)
	requesterNode := createNodeWithPeersRatingHandler(core.MetachainShardId, numOfShards, peersRatingCfg)

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

	hdr, hdrHash := getHeader()

	numOfRequests := 10
	// Add header to the resolver node's cache
	resolverNode.DataPool.Headers().AddHeader(hdrHash, hdr)
	requestHeader(requesterNode, numOfRequests, hdrHash, resolverNode.ShardCoordinator.SelfId())

	peerRatingsMap := getRatingsMapFromMetric(t, requesterNode)
	// resolver node should have received and responded to numOfRequests
	initialResolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialResolverExpectedRating := numOfRequests * (decreaseFactor + increaseFactor)
	assert.Equal(t, int32(initialResolverExpectedRating), initialResolverRating.Rating)
	testTimestampsForRespondingNode(t, initialResolverRating.TimestampLastResponseFromPid, initialResolverRating.TimestampLastRequestToPid)
	// malicious node should have only received numOfRequests
	initialMaliciousRating, exists := peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	initialMaliciousExpectedRating := numOfRequests * decreaseFactor
	assert.Equal(t, int32(initialMaliciousExpectedRating), initialMaliciousRating.Rating)
	testTimestampsForNotRespondingNode(t, initialMaliciousRating.TimestampLastResponseFromPid, initialMaliciousRating.TimestampLastRequestToPid, 0)

	maliciousNode.Close()

	// sleep enough so malicious node gets removed
	time.Sleep(time.Second * 15)
	peerRatingsMap = getRatingsMapFromMetric(t, requesterNode)
	_, exists = peerRatingsMap[maliciousNode.Messenger.ID().Pretty()]
	require.False(t, exists)
	resolverRating, exists := peerRatingsMap[resolverNode.Messenger.ID().Pretty()]
	require.True(t, exists)
	assert.Equal(t, initialResolverRating, resolverRating)
}

func createNodeWithPeersRatingHandler(shardID uint32, numShards uint32, cfg config.PeersRatingConfig) *integrationTests.TestProcessorNode {
	statusMetrics := statusHandler.NewStatusMetrics()
	appStatusHandler := &statusHandlerMock.AppStatusHandlerStub{
		SetStringValueHandler: func(key string, value string) {
			statusMetrics.SetStringValue(key, value)
		},
	}
	peersRatingHandler, _ := p2pFactory.NewPeersRatingHandler(
		p2pFactory.ArgPeersRatingHandler{
			TopRatedCache:              testscommon.NewCacherMock(),
			BadRatedCache:              testscommon.NewCacherMock(),
			MarkedForRemovalCache:      testscommon.NewCacherMock(),
			AppStatusHandler:           appStatusHandler,
			TimeWaitingForReconnection: time.Duration(cfg.TimeWaitingForReconnectionInSec) * time.Second,
			TimeBetweenMetricsUpdate:   time.Duration(cfg.TimeBetweenMetricsUpdateInSec) * time.Second,
			TimeBetweenCachersSweep:    time.Duration(cfg.TimeBetweenCachersSweepInSec) * time.Second,
		})

	return integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:          numShards,
		NodeShardId:        shardID,
		AppStatusHandler:   appStatusHandler,
		StatusMetrics:      statusMetrics,
		PeersRatingHandler: peersRatingHandler,
	})
}

func getHeader() (*block.Header, []byte) {
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
	}
	hdrBuff, _ := integrationTests.TestMarshalizer.Marshal(hdr)
	hdrHash := integrationTests.TestHasher.Compute(string(hdrBuff))
	return hdr, hdrHash
}

func getRatingsMapFromMetric(t *testing.T, node *integrationTests.TestProcessorNode) map[string]*ratingInfo {
	statusMetrics := node.Node.GetStatusCoreComponents().StatusMetrics()
	p2pMetricsMap, err := statusMetrics.StatusP2pMetricsMap()
	require.Nil(t, err)

	metricPeersRating := p2pMetricsMap[common.MetricP2PPeersRating]
	metricPeersRatingString, ok := metricPeersRating.(string)
	require.True(t, ok)

	peerRatingsMap := map[string]*ratingInfo{}
	err = json.Unmarshal([]byte(metricPeersRatingString), &peerRatingsMap)
	require.Nil(t, err)

	return peerRatingsMap
}

func requestHeader(requesterNode *integrationTests.TestProcessorNode, numOfRequests int, hdrHash []byte, shardID uint32) {
	for i := 0; i < numOfRequests; i++ {
		requesterNode.RequestHandler.RequestShardHeader(shardID, hdrHash)
		time.Sleep(time.Second) // allow nodes to respond
	}
}

func testTimestampsForRespondingNode(t *testing.T, timestampLastResponse int64, timestampLastRequest int64) {
	expectedMaxTimestamp := time.Now().Unix()
	expectedMinTimestamp := time.Now().Unix() - 1
	assert.LessOrEqual(t, timestampLastRequest, expectedMaxTimestamp)
	assert.GreaterOrEqual(t, timestampLastRequest, expectedMinTimestamp)
	assert.LessOrEqual(t, timestampLastResponse, expectedMaxTimestamp)
	assert.GreaterOrEqual(t, timestampLastResponse, expectedMinTimestamp)
}

func testTimestampsForNotRespondingNode(t *testing.T, timestampLastResponse int64, timestampLastRequest int64, expectedTimestampLastResponse int64) {
	expectedMaxTimestamp := time.Now().Unix()
	expectedMinTimestamp := time.Now().Unix() - 1
	assert.LessOrEqual(t, timestampLastRequest, expectedMaxTimestamp)
	assert.GreaterOrEqual(t, timestampLastRequest, expectedMinTimestamp)
	assert.Equal(t, expectedTimestampLastResponse, timestampLastResponse)
}
