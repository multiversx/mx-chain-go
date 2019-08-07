package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// topicRequestSuffix represents the topic name suffix
const topicRequestSuffix = "_REQUEST"

// NumPeersToQuery number of peers to send the message
const NumPeersToQuery = 2

type topicResolverSender struct {
	messenger       dataRetriever.MessageHandler
	marshalizer     marshal.Marshalizer
	topicName       string
	peerListCreator dataRetriever.PeerListCreator
	randomizer      dataRetriever.IntRandomizer
	targetShardId   uint32
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(
	messenger dataRetriever.MessageHandler,
	topicName string,
	peerListCreator dataRetriever.PeerListCreator,
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
	if peerListCreator == nil {
		return nil, dataRetriever.ErrNilPeerListCreator
	}

	resolver := &topicResolverSender{
		messenger:       messenger,
		topicName:       topicName,
		peerListCreator: peerListCreator,
		marshalizer:     marshalizer,
		randomizer:      randomizer,
		targetShardId:   targetShardId,
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

	peerList := trs.peerListCreator.PeerList()
	if len(peerList) == 0 {
		return dataRetriever.ErrNoConnectedPeerToSendRequest
	}

	topicToSendRequest := trs.topicName + topicRequestSuffix

	indexes := createIndexList(len(peerList))
	shuffledIndexes, err := fisherYatesShuffle(indexes, trs.randomizer)
	if err != nil {
		return err
	}

	msgSentCounter := 0
	for idx := range shuffledIndexes {
		peer := peerList[idx]

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
