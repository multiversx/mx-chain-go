package topicsender

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

var _ dataRetriever.TopicRequestSender = (*topicRequestSender)(nil)

const (
	shardBlocksTopicPrefix = "shardBlocks"
	// Keep below the resolver-side shardBlocks* topic antiflood limit of 30 messages/sec.
	maxShardBlockRequestsPerPeerPerSecond = 25
	shardBlockRequestRateLimitWindow      = time.Second
	requestRateLimiterMapKeySeparator     = "|"
)

// ArgTopicRequestSender is the argument structure used to create new topic request sender instance
type ArgTopicRequestSender struct {
	ArgBaseTopicSender
	Marshaller                  marshal.Marshalizer
	Randomizer                  dataRetriever.IntRandomizer
	PeerListCreator             dataRetriever.PeerListCreator
	NumIntraShardPeers          int
	NumCrossShardPeers          int
	NumFullHistoryPeers         int
	CurrentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	SelfShardIdProvider         dataRetriever.SelfShardIDProvider
	PeersRatingHandler          dataRetriever.PeersRatingHandler
}

type topicRequestSender struct {
	*baseTopicSender
	marshaller                         marshal.Marshalizer
	peerListCreator                    dataRetriever.PeerListCreator
	randomizer                         dataRetriever.IntRandomizer
	mutNumPeersToQuery                 sync.RWMutex
	numIntraShardPeers                 int
	numCrossShardPeers                 int
	numFullHistoryPeers                int
	currentNetworkEpochProviderHandler dataRetriever.CurrentNetworkEpochProviderHandler
	peersRatingHandler                 dataRetriever.PeersRatingHandler
	selfShardId                        uint32
	requestRateLimiter                 *requestRateLimiter
}

type sendOnTopicResult struct {
	numSent      int
	numThrottled int
}

type requestRateLimiter struct {
	mut      sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
}

// NewTopicRequestSender returns a new topic request sender instance
func NewTopicRequestSender(args ArgTopicRequestSender) (*topicRequestSender, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &topicRequestSender{
		baseTopicSender:                    createBaseTopicSender(args.ArgBaseTopicSender),
		marshaller:                         args.Marshaller,
		peerListCreator:                    args.PeerListCreator,
		randomizer:                         args.Randomizer,
		numIntraShardPeers:                 args.NumIntraShardPeers,
		numCrossShardPeers:                 args.NumCrossShardPeers,
		numFullHistoryPeers:                args.NumFullHistoryPeers,
		currentNetworkEpochProviderHandler: args.CurrentNetworkEpochProvider,
		peersRatingHandler:                 args.PeersRatingHandler,
		selfShardId:                        args.SelfShardIdProvider.SelfId(),
		requestRateLimiter:                 newRequestRateLimiter(maxShardBlockRequestsPerPeerPerSecond, shardBlockRequestRateLimitWindow),
	}, nil
}

func newRequestRateLimiter(limit int, window time.Duration) *requestRateLimiter {
	return &requestRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		now:      time.Now,
	}
}

func (rl *requestRateLimiter) allow(topic string, peer core.PeerID) bool {
	rl.mut.Lock()
	defer rl.mut.Unlock()

	now := rl.now()
	cutoff := now.Add(-rl.window)
	key := string(peer) + requestRateLimiterMapKeySeparator + topic
	requests := rl.requests[key]

	firstValidIndex := 0
	for firstValidIndex < len(requests) && !requests[firstValidIndex].After(cutoff) {
		firstValidIndex++
	}
	if firstValidIndex > 0 {
		requests = requests[firstValidIndex:]
	}

	if len(requests) >= rl.limit {
		rl.requests[key] = requests
		return false
	}

	requests = append(requests, now)
	rl.requests[key] = requests
	return true
}

func checkArgs(args ArgTopicRequestSender) error {
	err := checkBaseTopicSenderArgs(args.ArgBaseTopicSender)
	if err != nil {
		return err
	}
	if check.IfNil(args.Marshaller) {
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.Randomizer) {
		return dataRetriever.ErrNilRandomizer
	}
	if check.IfNil(args.PeerListCreator) {
		return dataRetriever.ErrNilPeerListCreator
	}
	if check.IfNil(args.CurrentNetworkEpochProvider) {
		return dataRetriever.ErrNilCurrentNetworkEpochProvider
	}
	if check.IfNil(args.PeersRatingHandler) {
		return dataRetriever.ErrNilPeersRatingHandler
	}
	if check.IfNil(args.SelfShardIdProvider) {
		return dataRetriever.ErrNilSelfShardIDProvider
	}
	if args.NumIntraShardPeers < 0 {
		return fmt.Errorf("%w for NumIntraShardPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if args.NumCrossShardPeers < 0 {
		return fmt.Errorf("%w for NumCrossShardPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if args.NumFullHistoryPeers < 0 {
		return fmt.Errorf("%w for NumFullHistoryPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if args.NumCrossShardPeers+args.NumIntraShardPeers < minPeersToQuery {
		return fmt.Errorf("%w for NumCrossShardPeers, NumIntraShardPeers as their sum should be greater or equal than %d",
			dataRetriever.ErrInvalidValue, minPeersToQuery)
	}
	return nil
}

// SendOnRequestTopic is used to send request data over channels (topics) to other peers
// This method only sends the request, the received data should be handled by interceptors
func (trs *topicRequestSender) SendOnRequestTopic(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
	buff, err := trs.marshaller.Marshal(rd)
	if err != nil {
		return err
	}

	topicToSendRequest := trs.topicName + core.TopicRequestSuffix

	var numSentIntra, numSentCross int
	var numThrottledFullHistory int
	var intraPeers, crossPeers []core.PeerID
	fullHistoryPeers := make([]core.PeerID, 0)
	requestedNetworks := make([]string, 0)
	if !trs.currentNetworkEpochProviderHandler.EpochIsActiveInNetwork(rd.Epoch) {
		preferredPeer := trs.getPreferredFullArchivePeer()
		fullHistoryPeers = trs.fullArchiveMessenger.ConnectedPeers()

		result := trs.sendOnTopic(
			fullHistoryPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numFullHistoryPeers,
			core.FullHistoryPeer.String(),
			trs.fullArchiveMessenger)
		numSentIntra = result.numSent
		numThrottledFullHistory = result.numThrottled

		requestedNetworks = append(requestedNetworks, "full archive network")
	}

	if numSentCross+numSentIntra == 0 && numThrottledFullHistory == 0 {
		crossPeers = trs.peerListCreator.CrossShardPeerList()
		preferredPeer := trs.getPreferredPeer(trs.targetShardId)
		result := trs.sendOnTopic(
			crossPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numCrossShardPeers,
			core.CrossShardPeer.String(),
			trs.mainMessenger)
		numSentCross = result.numSent

		intraPeers = trs.peerListCreator.IntraShardPeerList()
		preferredPeer = trs.getPreferredPeer(trs.selfShardId)
		result = trs.sendOnTopic(
			intraPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numIntraShardPeers,
			core.IntraShardPeer.String(),
			trs.mainMessenger)
		numSentIntra = result.numSent

		requestedNetworks = append(requestedNetworks, "main network")
	}

	trs.callDebugHandler(originalHashes, numSentIntra, numSentCross)

	if numSentCross+numSentIntra == 0 {
		return fmt.Errorf("%w, topic: %s, crossPeers: %d, intraPeers: %d, fullHistoryPeers: %d, requested networks: %s",
			dataRetriever.ErrSendRequest,
			trs.topicName,
			len(crossPeers),
			len(intraPeers),
			len(fullHistoryPeers),
			strings.Join(requestedNetworks, ", "),
		)
	}

	return nil
}

func (trs *topicRequestSender) callDebugHandler(originalHashes [][]byte, numSentIntra int, numSentCross int) {
	trs.mutDebugHandler.RLock()
	defer trs.mutDebugHandler.RUnlock()

	trs.debugHandler.LogRequestedData(trs.topicName, originalHashes, numSentIntra, numSentCross)
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}

func (trs *topicRequestSender) sendOnTopic(
	peerList []core.PeerID,
	preferredPeer core.PeerID,
	topicToSendRequest string,
	buff []byte,
	maxToSend int,
	peerType string,
	messenger p2p.MessageHandler,
) sendOnTopicResult {
	if len(peerList) == 0 || maxToSend == 0 {
		return sendOnTopicResult{}
	}

	histogramMap := make(map[string]int)

	topRatedPeersList := trs.peersRatingHandler.GetTopRatedPeersFromList(peerList, len(peerList))

	indexes := createIndexList(len(topRatedPeersList))
	shuffledIndexes := random.FisherYatesShuffle(indexes, trs.randomizer)
	logData := make([]interface{}, 0)
	msgSentCounter := 0
	numThrottled := 0
	shouldSendToPreferredPeer := preferredPeer != "" && maxToSend > 1
	if shouldSendToPreferredPeer {
		shuffledIndexes = append([]int{preferredPeerIndex}, shuffledIndexes...)
	}

	for idx := 0; idx < len(shuffledIndexes); idx++ {
		peer := getPeerID(shuffledIndexes[idx], topRatedPeersList, preferredPeer, peerType, topicToSendRequest, histogramMap)

		if !trs.canSendRequestToPeer(topicToSendRequest, peer) {
			numThrottled++
			log.Debug("request to peer throttled",
				"topic", topicToSendRequest,
				"peer", peer.Pretty(),
				"peer type", peerType,
			)
			continue
		}

		// no matter the outcome of sendToConnectedPeer, decrease the peer's rating
		// this way we avoid(decreasing) peers with invalid connections or blacklisted by antiflooder
		trs.peersRatingHandler.DecreaseRating(peer)

		err := trs.sendToConnectedPeer(topicToSendRequest, buff, peer, messenger)
		if err != nil {
			log.Trace("sendToConnectedPeer failed",
				"topic", topicToSendRequest,
				"peer", peer.Pretty(),
				"peer type", peerType,
				"error", err.Error())
			continue
		}

		logData = append(logData, peerType)
		logData = append(logData, peer.Pretty())
		msgSentCounter++
		if msgSentCounter == maxToSend {
			break
		}
	}
	log.Trace("requests are sent to", logData...)
	if numThrottled > 0 {
		histogramMap["throttled"] = numThrottled
	}
	log.Trace("request peers histogram", "max peers to send", maxToSend, "topic", topicToSendRequest, "histogram", histogramMap)

	return sendOnTopicResult{
		numSent:      msgSentCounter,
		numThrottled: numThrottled,
	}
}

func (trs *topicRequestSender) canSendRequestToPeer(topic string, peer core.PeerID) bool {
	if !isShardBlocksRequestTopic(topic) {
		return true
	}

	return trs.requestRateLimiter.allow(topic, peer)
}

func isShardBlocksRequestTopic(topic string) bool {
	return strings.HasPrefix(topic, shardBlocksTopicPrefix) && strings.HasSuffix(topic, core.TopicRequestSuffix)
}

func getPeerID(index int, peersList []core.PeerID, preferredPeer core.PeerID, peerType string, topic string, histogramMap map[string]int) core.PeerID {
	if index == preferredPeerIndex {
		histogramMap["preferred"]++
		log.Trace("sending request to preferred peer", "peer", preferredPeer.Pretty(), "topic", topic, "peer type", peerType)

		return preferredPeer
	}

	histogramMap[peerType]++
	return peersList[index]
}

func (trs *topicRequestSender) getPreferredPeer(shardID uint32) core.PeerID {
	peersInShard, found := trs.getPreferredPeersInShard(shardID)
	if !found {
		return ""
	}

	return trs.getRandomPeerID(peersInShard)
}

func (trs *topicRequestSender) getPreferredPeersInShard(shardID uint32) ([]core.PeerID, bool) {
	preferredPeers := trs.mainPreferredPeersHolderHandler.Get()

	peers, found := preferredPeers[shardID]
	if !found || len(peers) == 0 {
		return nil, false
	}

	return peers, true
}

func (trs *topicRequestSender) getPreferredFullArchivePeer() core.PeerID {
	preferredPeersMap := trs.fullArchivePreferredPeersHolderHandler.Get()
	preferredPeersSlice := mapToSlice(preferredPeersMap)

	if len(preferredPeersSlice) == 0 {
		return ""
	}

	return trs.getRandomPeerID(preferredPeersSlice)
}

func (trs *topicRequestSender) getRandomPeerID(peerIDs []core.PeerID) core.PeerID {
	randomIdx := trs.randomizer.Intn(len(peerIDs))

	return peerIDs[randomIdx]
}

func mapToSlice(initialMap map[uint32][]core.PeerID) []core.PeerID {
	newSlice := make([]core.PeerID, 0, len(initialMap))

	for _, peerIDsOnShard := range initialMap {
		newSlice = append(newSlice, peerIDsOnShard...)
	}

	return newSlice
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peers to query
func (trs *topicRequestSender) SetNumPeersToQuery(intra int, cross int) {
	trs.mutNumPeersToQuery.Lock()
	trs.numIntraShardPeers = intra
	trs.numCrossShardPeers = cross
	trs.mutNumPeersToQuery.Unlock()
}

// NumPeersToQuery will return the number of intra shard and cross shard number of peer to query
func (trs *topicRequestSender) NumPeersToQuery() (int, int) {
	trs.mutNumPeersToQuery.RLock()
	defer trs.mutNumPeersToQuery.RUnlock()

	return trs.numIntraShardPeers, trs.numCrossShardPeers
}

// IsInterfaceNil returns true if there is no value under the interface
func (trs *topicRequestSender) IsInterfaceNil() bool {
	return trs == nil
}
