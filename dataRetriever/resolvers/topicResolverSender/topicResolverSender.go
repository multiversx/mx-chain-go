package topicResolverSender

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// topicRequestSuffix represents the topic name suffix
const topicRequestSuffix = "_REQUEST"

// NumPeersToQuery number of peers to send the message
const NumPeersToQuery = 2

var log = logger.DefaultLogger()

type topicResolverSender struct {
	messenger   dataRetriever.MessageHandler
	marshalizer marshal.Marshalizer
	topicName   string
	randomizer  dataRetriever.IntRandomizer
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(
	messenger dataRetriever.MessageHandler,
	topicName string,
	marshalizer marshal.Marshalizer,
	randomizer dataRetriever.IntRandomizer,
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
		messenger:   messenger,
		topicName:   topicName,
		marshalizer: marshalizer,
		randomizer:  randomizer,
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
	peersToSend, err := selectRandomPeers(trs.messenger.ConnectedPeersOnTopic(topicToSendRequest), NumPeersToQuery, trs.randomizer)
	if err != nil {
		return err
	}
	if len(peersToSend) == 0 {
		return dataRetriever.ErrNoConnectedPeerToSendRequest
	}

	messageSent := false
	for _, peer := range peersToSend {
		err = trs.messenger.SendToConnectedPeer(topicToSendRequest, buff, peer)
		if err != nil {
			log.Debug(err.Error())
		} else {
			messageSent = true
		}
	}

	if !messageSent {
		return err
	}

	return nil
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

func selectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer dataRetriever.IntRandomizer) ([]p2p.PeerID, error) {
	selectedPeers := make([]p2p.PeerID, 0)

	if len(connectedPeers) == 0 {
		return selectedPeers, nil
	}

	if len(connectedPeers) <= peersToSend {
		return connectedPeers, nil
	}

	uniqueIndexes := make(map[int]struct{})
	//generating peersToSend number of unique indexes
	for len(uniqueIndexes) < peersToSend {
		newIndex, err := randomizer.Intn(len(connectedPeers))
		if err != nil {
			return nil, err
		}

		uniqueIndexes[newIndex] = struct{}{}
	}

	for index := range uniqueIndexes {
		selectedPeers = append(selectedPeers, connectedPeers[index])
	}

	return selectedPeers, nil
}
