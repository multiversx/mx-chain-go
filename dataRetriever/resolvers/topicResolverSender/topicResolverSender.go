package topicResolverSender

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/random"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	resolverDebug "github.com/ElrondNetwork/elrond-go/debug/resolver"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
)

// topicRequestSuffix represents the topic name suffix
const topicRequestSuffix = "_REQUEST"

const minPeersToQuery = 2

const preferredPeerIndex = -1

type peerType string

func (pt peerType) String() string {
	return string(pt)
}

const (
	intraShardPeer  peerType = "intra peer"
	crossShardPeer  peerType = "cross peer"
	fullHistoryPeer peerType = "full history peer"
)

var _ dataRetriever.TopicResolverSender = (*topicResolverSender)(nil)
var log = logger.GetOrCreate("dataretriever/resolverstopicresolversender")

// ArgTopicResolverSender is the argument structure used to create new TopicResolverSender instance
type ArgTopicResolverSender struct {
	Messenger                   dataRetriever.MessageHandler
	TopicName                   string
	PeerListCreator             dataRetriever.PeerListCreator
	Marshalizer                 marshal.Marshalizer
	Randomizer                  dataRetriever.IntRandomizer
	OutputAntiflooder           dataRetriever.P2PAntifloodHandler
	NumIntraShardPeers          int
	NumCrossShardPeers          int
	NumFullHistoryPeers         int
	CurrentNetworkEpochProvider dataRetriever.CurrentNetworkEpochProviderHandler
	PreferredPeersHolder        dataRetriever.PreferredPeersHolderHandler
	SelfShardIdProvider         dataRetriever.SelfShardIDProvider
	TargetShardId               uint32
}

type topicResolverSender struct {
	messenger                          dataRetriever.MessageHandler
	marshalizer                        marshal.Marshalizer
	topicName                          string
	peerListCreator                    dataRetriever.PeerListCreator
	randomizer                         dataRetriever.IntRandomizer
	outputAntiflooder                  dataRetriever.P2PAntifloodHandler
	mutNumPeersToQuery                 sync.RWMutex
	numIntraShardPeers                 int
	numCrossShardPeers                 int
	numFullHistoryPeers                int
	mutResolverDebugHandler            sync.RWMutex
	resolverDebugHandler               dataRetriever.ResolverDebugHandler
	currentNetworkEpochProviderHandler dataRetriever.CurrentNetworkEpochProviderHandler
	preferredPeersHolderHandler        dataRetriever.PreferredPeersHolderHandler
	selfShardId                        uint32
	targetShardId                      uint32
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(arg ArgTopicResolverSender) (*topicResolverSender, error) {
	if check.IfNil(arg.Messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.Randomizer) {
		return nil, dataRetriever.ErrNilRandomizer
	}
	if check.IfNil(arg.PeerListCreator) {
		return nil, dataRetriever.ErrNilPeerListCreator
	}
	if check.IfNil(arg.OutputAntiflooder) {
		return nil, dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.CurrentNetworkEpochProvider) {
		return nil, dataRetriever.ErrNilCurrentNetworkEpochProvider
	}
	if check.IfNil(arg.PreferredPeersHolder) {
		return nil, dataRetriever.ErrNilPreferredPeersHolder
	}
	if check.IfNil(arg.SelfShardIdProvider) {
		return nil, dataRetriever.ErrNilSelfShardIDProvider
	}
	if arg.NumIntraShardPeers < 0 {
		return nil, fmt.Errorf("%w for NumIntraShardPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if arg.NumCrossShardPeers < 0 {
		return nil, fmt.Errorf("%w for NumCrossShardPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if arg.NumFullHistoryPeers < 0 {
		return nil, fmt.Errorf("%w for NumFullHistoryPeers as the value should be greater or equal than 0",
			dataRetriever.ErrInvalidValue)
	}
	if arg.NumCrossShardPeers+arg.NumIntraShardPeers < minPeersToQuery {
		return nil, fmt.Errorf("%w for NumCrossShardPeers, NumIntraShardPeers as their sum should be greater or equal than %d",
			dataRetriever.ErrInvalidValue, minPeersToQuery)
	}

	resolver := &topicResolverSender{
		messenger:                          arg.Messenger,
		topicName:                          arg.TopicName,
		peerListCreator:                    arg.PeerListCreator,
		marshalizer:                        arg.Marshalizer,
		randomizer:                         arg.Randomizer,
		targetShardId:                      arg.TargetShardId,
		selfShardId:                        arg.SelfShardIdProvider.SelfId(),
		outputAntiflooder:                  arg.OutputAntiflooder,
		numIntraShardPeers:                 arg.NumIntraShardPeers,
		numCrossShardPeers:                 arg.NumCrossShardPeers,
		numFullHistoryPeers:                arg.NumFullHistoryPeers,
		currentNetworkEpochProviderHandler: arg.CurrentNetworkEpochProvider,
		preferredPeersHolderHandler:        arg.PreferredPeersHolder,
	}
	resolver.resolverDebugHandler = resolverDebug.NewDisabledInterceptorResolver()

	return resolver, nil
}

// SendOnRequestTopic is used to send request data over channels (topics) to other peers
// This method only sends the request, the received data should be handled by interceptors
func (trs *topicResolverSender) SendOnRequestTopic(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
	buff, err := trs.marshalizer.Marshal(rd)
	if err != nil {
		return err
	}

	topicToSendRequest := trs.topicName + topicRequestSuffix

	var numSentIntra, numSentCross int
	var intraPeers, crossPeers []core.PeerID
	fullHistoryPeers := make([]core.PeerID, 0)
	if trs.currentNetworkEpochProviderHandler.EpochIsActiveInNetwork(rd.Epoch) {
		crossPeers = trs.peerListCreator.CrossShardPeerList()
		preferredPeer := trs.getPreferredPeer(crossShardPeer)
		numSentCross = trs.sendOnTopic(crossPeers, preferredPeer, topicToSendRequest, buff, trs.numCrossShardPeers, crossShardPeer.String())

		intraPeers = trs.peerListCreator.IntraShardPeerList()
		preferredPeer = trs.getPreferredPeer(intraShardPeer)
		numSentIntra = trs.sendOnTopic(intraPeers, preferredPeer, topicToSendRequest, buff, trs.numIntraShardPeers, intraShardPeer.String())
	} else {
		// TODO: select preferred peers of type full history as well.
		fullHistoryPeers = trs.peerListCreator.FullHistoryList()
		numSentIntra = trs.sendOnTopic(fullHistoryPeers, "", topicToSendRequest, buff, trs.numFullHistoryPeers, fullHistoryPeer.String())
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

func (trs *topicResolverSender) callDebugHandler(originalHashes [][]byte, numSentIntra int, numSentCross int) {
	trs.mutResolverDebugHandler.RLock()
	defer trs.mutResolverDebugHandler.RUnlock()

	trs.resolverDebugHandler.LogRequestedData(trs.topicName, originalHashes, numSentIntra, numSentCross)
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}

func (trs *topicResolverSender) sendOnTopic(
	peerList []core.PeerID,
	preferredPeer core.PeerID,
	topicToSendRequest string,
	buff []byte,
	maxToSend int,
	peerType string,
) int {
	if len(peerList) == 0 || maxToSend == 0 {
		return 0
	}

	// TODO: remove this map before merging or move to level TRACE
	histogramMap := make(map[string]int)

	indexes := createIndexList(len(peerList))
	shuffledIndexes := random.FisherYatesShuffle(indexes, trs.randomizer)
	logData := make([]interface{}, 0)
	msgSentCounter := 0
	shouldSendToPreferredPeer := preferredPeer != "" && maxToSend > 1
	if shouldSendToPreferredPeer {
		shuffledIndexes = append([]int{preferredPeerIndex}, shuffledIndexes...)
	}

	for idx := 0; idx < len(shuffledIndexes); idx++ {
		peer := getPeerID(shuffledIndexes[idx], peerList, preferredPeer, peerType, topicToSendRequest, histogramMap)

		err := trs.sendToConnectedPeer(topicToSendRequest, buff, peer)
		if err != nil {
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

	log.Debug("request peers histogram", "max peers to send", maxToSend, "topic", topicToSendRequest, "histogram", histogramMap)

	return msgSentCounter
}

func getPeerID(index int, peersList []core.PeerID, preferredPeer core.PeerID, peerType string, topic string, histogramMap map[string]int) core.PeerID {
	if index == preferredPeerIndex {
		histogramMap["preferred"]++
		log.Debug("sending request to preferred peer", "peer", preferredPeer.Pretty(), "topic", topic, "peer type", peerType)

		return preferredPeer
	}

	histogramMap[peerType]++
	return peersList[index]
}

func (trs *topicResolverSender) getPreferredPeer(pType peerType) core.PeerID {
	peersInShard, found := trs.getPeersBasedOnPeerType(pType)
	if !found {
		return ""
	}

	randomIdx := trs.randomizer.Intn(len(peersInShard))

	return peersInShard[randomIdx]
}

func (trs *topicResolverSender) getPeersBasedOnPeerType(pType peerType) ([]core.PeerID, bool) {
	preferredPeers := trs.preferredPeersHolderHandler.Get()

	var shardToSearch uint32
	if pType == crossShardPeer {
		shardToSearch = trs.targetShardId
	} else {
		shardToSearch = trs.selfShardId
	}

	peers, found := preferredPeers[shardToSearch]
	if !found || len(peers) == 0 {
		return nil, false
	}

	return peers, true
}

// Send is used to send an array buffer to a connected peer
// It is used when replying to a request
func (trs *topicResolverSender) Send(buff []byte, peer core.PeerID) error {
	return trs.sendToConnectedPeer(trs.topicName, buff, peer)
}

func (trs *topicResolverSender) sendToConnectedPeer(topic string, buff []byte, peer core.PeerID) error {
	msg := &message.Message{
		DataField:  buff,
		PeerField:  peer,
		TopicField: topic,
	}

	err := trs.outputAntiflooder.CanProcessMessage(msg, peer)
	if err != nil {
		return fmt.Errorf("%w while sending %d bytes to peer %s",
			err,
			len(buff),
			p2p.PeerIdToShortString(peer),
		)
	}

	return trs.messenger.SendToConnectedPeer(topic, buff, peer)
}

// ResolverDebugHandler returns the debug handler used in resolvers
func (trs *topicResolverSender) ResolverDebugHandler() dataRetriever.ResolverDebugHandler {
	trs.mutResolverDebugHandler.RLock()
	defer trs.mutResolverDebugHandler.RUnlock()

	return trs.resolverDebugHandler
}

// SetResolverDebugHandler sets the debug handler used in resolvers
func (trs *topicResolverSender) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if check.IfNil(handler) {
		return dataRetriever.ErrNilResolverDebugHandler
	}

	trs.mutResolverDebugHandler.Lock()
	trs.resolverDebugHandler = handler
	trs.mutResolverDebugHandler.Unlock()

	return nil
}

// RequestTopic returns the topic with the request suffix used for sending requests
func (trs *topicResolverSender) RequestTopic() string {
	return trs.topicName + topicRequestSuffix
}

// TargetShardID returns the target shard ID for this resolver should serve data
func (trs *topicResolverSender) TargetShardID() uint32 {
	return trs.targetShardId
}

// SetNumPeersToQuery will set the number of intra shard and cross shard number of peers to query
func (trs *topicResolverSender) SetNumPeersToQuery(intra int, cross int) {
	trs.mutNumPeersToQuery.Lock()
	trs.numIntraShardPeers = intra
	trs.numCrossShardPeers = cross
	trs.mutNumPeersToQuery.Unlock()
}

// NumPeersToQuery will return the number of intra shard and cross shard number of peer to query
func (trs *topicResolverSender) NumPeersToQuery() (int, int) {
	trs.mutNumPeersToQuery.RLock()
	defer trs.mutNumPeersToQuery.RUnlock()

	return trs.numIntraShardPeers, trs.numCrossShardPeers
}

// IsInterfaceNil returns true if there is no value under the interface
func (trs *topicResolverSender) IsInterfaceNil() bool {
	return trs == nil
}
