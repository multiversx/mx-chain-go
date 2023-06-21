package topicsender

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

var _ dataRetriever.TopicRequestSender = (*topicRequestSender)(nil)

// ArgTopicRequestSender is the argument structure used to create new topic request sender instance
type ArgTopicRequestSender struct {
	ArgBaseTopicSender
	Marshaller                    marshal.Marshalizer
	Randomizer                    dataRetriever.IntRandomizer
	PeerListCreator               dataRetriever.PeerListCreator
	NumIntraShardPeers            int
	NumCrossShardPeers            int
	NumFullHistoryPeers           int
	CurrentNetworkEpochProvider   dataRetriever.CurrentNetworkEpochProviderHandler
	SelfShardIdProvider           dataRetriever.SelfShardIDProvider
	MainPeersRatingHandler        dataRetriever.PeersRatingHandler
	FullArchivePeersRatingHandler dataRetriever.PeersRatingHandler
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
	mainPeersRatingHandler             dataRetriever.PeersRatingHandler
	fullArchivePeersRatingHandler      dataRetriever.PeersRatingHandler
	selfShardId                        uint32
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
		mainPeersRatingHandler:             args.MainPeersRatingHandler,
		fullArchivePeersRatingHandler:      args.FullArchivePeersRatingHandler,
		selfShardId:                        args.SelfShardIdProvider.SelfId(),
	}, nil
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
	if check.IfNil(args.MainPeersRatingHandler) {
		return fmt.Errorf("%w on main network", dataRetriever.ErrNilPeersRatingHandler)
	}
	if check.IfNil(args.FullArchivePeersRatingHandler) {
		return fmt.Errorf("%w on full archive network", dataRetriever.ErrNilPeersRatingHandler)
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
	var intraPeers, crossPeers []core.PeerID
	fullHistoryPeers := make([]core.PeerID, 0)
	if trs.currentNetworkEpochProviderHandler.EpochIsActiveInNetwork(rd.Epoch) {
		crossPeers = trs.peerListCreator.CrossShardPeerList()
		preferredPeer := trs.getPreferredPeer(trs.targetShardId)
		numSentCross = trs.sendOnTopic(
			crossPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numCrossShardPeers,
			core.CrossShardPeer.String(),
			trs.mainMessenger,
			trs.mainPeersRatingHandler,
			p2p.MainNetwork,
			trs.mainPreferredPeersHolderHandler)

		intraPeers = trs.peerListCreator.IntraShardPeerList()
		preferredPeer = trs.getPreferredPeer(trs.selfShardId)
		numSentIntra = trs.sendOnTopic(
			intraPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numIntraShardPeers,
			core.IntraShardPeer.String(),
			trs.mainMessenger,
			trs.mainPeersRatingHandler,
			p2p.MainNetwork,
			trs.mainPreferredPeersHolderHandler)
	} else {
		preferredPeer := trs.getPreferredFullArchivePeer()
		fullHistoryPeers = trs.fullArchiveMessenger.ConnectedPeers()

		numSentIntra = trs.sendOnTopic(
			fullHistoryPeers,
			preferredPeer,
			topicToSendRequest,
			buff,
			trs.numFullHistoryPeers,
			core.FullHistoryPeer.String(),
			trs.fullArchiveMessenger,
			trs.fullArchivePeersRatingHandler,
			p2p.FullArchiveNetwork,
			trs.fullArchivePreferredPeersHolderHandler)
	}

	trs.callDebugHandler(originalHashes, numSentIntra, numSentCross)

	if numSentCross+numSentIntra == 0 {
		return fmt.Errorf("%w, topic: %s, crossPeers: %d, intraPeers: %d, fullHistoryPeers: %d",
			dataRetriever.ErrSendRequest,
			trs.topicName,
			len(crossPeers),
			len(intraPeers),
			len(fullHistoryPeers))
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
	messenger dataRetriever.MessageHandler,
	peersRatingHandler dataRetriever.PeersRatingHandler,
	network p2p.Network,
	preferredPeersHolder dataRetriever.PreferredPeersHolderHandler,
) int {
	if len(peerList) == 0 || maxToSend == 0 {
		return 0
	}

	histogramMap := make(map[string]int)

	topRatedPeersList := peersRatingHandler.GetTopRatedPeersFromList(peerList, maxToSend)

	indexes := createIndexList(len(topRatedPeersList))
	shuffledIndexes := random.FisherYatesShuffle(indexes, trs.randomizer)
	logData := make([]interface{}, 0)
	msgSentCounter := 0
	shouldSendToPreferredPeer := preferredPeer != "" && maxToSend > 1
	if shouldSendToPreferredPeer {
		shuffledIndexes = append([]int{preferredPeerIndex}, shuffledIndexes...)
	}

	logData = append(logData, "network", network)

	for idx := 0; idx < len(shuffledIndexes); idx++ {
		peer := getPeerID(shuffledIndexes[idx], topRatedPeersList, preferredPeer, peerType, topicToSendRequest, histogramMap)

		err := trs.sendToConnectedPeer(topicToSendRequest, buff, peer, messenger, network, preferredPeersHolder)
		if err != nil {
			continue
		}
		peersRatingHandler.DecreaseRating(peer)

		logData = append(logData, peerType)
		logData = append(logData, peer.Pretty())
		msgSentCounter++
		if msgSentCounter == maxToSend {
			break
		}
	}
	log.Trace("requests are sent to", logData...)
	log.Trace("request peers histogram", "network", network, "max peers to send", maxToSend, "topic", topicToSendRequest, "histogram", histogramMap)

	return msgSentCounter
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
