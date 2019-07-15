package topicResolverSender

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// topicRequestSuffix represents the topic name suffix
const topicRequestSuffix = "_REQUEST"

// NumPeersToQuery number of peers to send the message
const NumPeersToQuery = 2

type topicResolverSender struct {
	messenger             dataRetriever.MessageHandler
	marshalizer           marshal.Marshalizer
	topicName             string
	excludePeersFromTopic string
	randomizer            dataRetriever.IntRandomizer
	targetShardId         uint32
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(
	messenger dataRetriever.MessageHandler,
	topicName string,
	excludePeersFromTopic string,
	marshalizer marshal.Marshalizer,
	randomizer dataRetriever.IntRandomizer,
	targetShardId uint32,
) (*topicResolverSender, error) {

	if messenger == nil {
		return nil, dataRetriever.ErrNilMessenger
	}
	if marshalizer == nil {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if randomizer == nil {
		return nil, dataRetriever.ErrNilRandomizer
	}

	resolver := &topicResolverSender{
		messenger:             messenger,
		topicName:             topicName,
		excludePeersFromTopic: excludePeersFromTopic,
		marshalizer:           marshalizer,
		randomizer:            randomizer,
		targetShardId:         targetShardId,
	}

	return resolver, nil
}

// SendOnRequestTopic is used to send request data over channels (topics) to other peers
// This method only sends the request, the received data should be handled by interceptors
func (trs *topicResolverSender) SendOnRequestTopic(rd *dataRetriever.RequestData) error {
	buff, err := trs.marshalizer.Marshal(rd)
	if err != nil {
		return err
	}

	topicToSendRequest := trs.topicName + topicRequestSuffix
	excludedTopicSendRequest := trs.excludePeersFromTopic + topicRequestSuffix
	allConnectedPeers := trs.messenger.ConnectedPeersOnTopic(topicToSendRequest)
	if len(allConnectedPeers) == 0 {
		return dataRetriever.ErrNoConnectedPeerToSendRequest
	}

	excludedConnectedPeers := make([]p2p.PeerID, 0)
	isExcludedTopicSet := len(trs.excludePeersFromTopic) > 0
	if isExcludedTopicSet {
		excludedConnectedPeers = trs.messenger.ConnectedPeersOnTopic(excludedTopicSendRequest)
	}

	diffList := makeDiffList(allConnectedPeers, excludedConnectedPeers)
	if len(diffList) == 0 {
		//no differences: fallback to all connected peers
		diffList = allConnectedPeers
	}

	indexes := createIndexList(len(diffList))
	shuffledIndexes, err := fisherYatesShuffle(indexes, trs.randomizer)
	if err != nil {
		return err
	}

	msgSentCounter := 0
	for idx := range shuffledIndexes {
		peer := diffList[idx]

		err = trs.messenger.SendToConnectedPeer(topicToSendRequest, buff, peer)
		if err != nil {
			continue
		}

		msgSentCounter++
		if msgSentCounter == NumPeersToQuery {
			break
		}
	}

	if msgSentCounter == 0 {
		return err
	}

	return nil
}

func createIndexList(listLength int) []int {
	indexes := make([]int, listLength)
	for i := 0; i < listLength; i++ {
		indexes[i] = i
	}

	return indexes
}

// Send is used to send an array buffer to a connected peer
// It is used when replying to a request
func (trs *topicResolverSender) Send(buff []byte, peer p2p.PeerID) error {
	return trs.messenger.SendToConnectedPeer(trs.topicName, buff, peer)
}

// TopicRequestSuffix returns the suffix that will be added to create a new channel for requests
func (trs *topicResolverSender) TopicRequestSuffix() string {
	return topicRequestSuffix
}

// TargetShardID returns the target shard ID for this resolver should serve data
func (trs *topicResolverSender) TargetShardID() uint32 {
	return trs.targetShardId
}

func fisherYatesShuffle(indexes []int, randomizer dataRetriever.IntRandomizer) ([]int, error) {
	newIndexes := make([]int, len(indexes))
	copy(newIndexes, indexes)

	for i := len(newIndexes) - 1; i > 0; i-- {
		j, err := randomizer.Intn(i + 1)
		if err != nil {
			return nil, err
		}

		newIndexes[i], newIndexes[j] = newIndexes[j], newIndexes[i]
	}

	return newIndexes, nil
}

func makeDiffList(
	allConnectedPeers []p2p.PeerID,
	excludedConnectedPeers []p2p.PeerID,
) []p2p.PeerID {

	if len(excludedConnectedPeers) == 0 {
		return allConnectedPeers
	}

	diff := make([]p2p.PeerID, 0)
	for _, pid := range allConnectedPeers {
		isPeerExcluded := false

		for _, excluded := range excludedConnectedPeers {
			if bytes.Equal(pid.Bytes(), excluded.Bytes()) {
				isPeerExcluded = true
				break
			}
		}

		if !isPeerExcluded {
			diff = append(diff, pid)
		}
	}

	return diff
}
