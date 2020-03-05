package topicResolverSender

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/message"
)

// topicRequestSuffix represents the topic name suffix
const topicRequestSuffix = "_REQUEST"

// NumPeersToQuery number of peers to send the message
const NumPeersToQuery = 2

// ArgTopicResolverSender is the argument structure used to create new TopicResolverSender instance
type ArgTopicResolverSender struct {
	Messenger         dataRetriever.MessageHandler
	TopicName         string
	PeerListCreator   dataRetriever.PeerListCreator
	Marshalizer       marshal.Marshalizer
	Randomizer        dataRetriever.IntRandomizer
	TargetShardId     uint32
	OutputAntiflooder dataRetriever.P2PAntifloodHandler
}

type topicResolverSender struct {
	messenger         dataRetriever.MessageHandler
	marshalizer       marshal.Marshalizer
	topicName         string
	peerListCreator   dataRetriever.PeerListCreator
	randomizer        dataRetriever.IntRandomizer
	targetShardId     uint32
	outputAntiflooder dataRetriever.P2PAntifloodHandler
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

	resolver := &topicResolverSender{
		messenger:         arg.Messenger,
		topicName:         arg.TopicName,
		peerListCreator:   arg.PeerListCreator,
		marshalizer:       arg.Marshalizer,
		randomizer:        arg.Randomizer,
		targetShardId:     arg.TargetShardId,
		outputAntiflooder: arg.OutputAntiflooder,
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

		err = trs.sendToConnectedPeer(topicToSendRequest, buff, peer)
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
	return trs.sendToConnectedPeer(trs.topicName, buff, peer)
}

func (trs *topicResolverSender) sendToConnectedPeer(topic string, buff []byte, peer p2p.PeerID) error {
	msg := &message.Message{
		DataField:     buff,
		PeerField:     peer,
		TopicIdsField: []string{topic},
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

// Topic returns the topic with the request suffix used for sending requests
func (trs *topicResolverSender) Topic() string {
	return trs.topicName + topicRequestSuffix
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

// IsInterfaceNil returns true if there is no value under the interface
func (trs *topicResolverSender) IsInterfaceNil() bool {
	return trs == nil
}
