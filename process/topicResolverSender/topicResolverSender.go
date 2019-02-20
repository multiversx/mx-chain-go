package topicResolverSender

import (
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// requestTopicSuffix represents the topic name suffix
const requestTopicSuffix = "_REQUEST"

// PeersToSendRequest number of peers to send the message
const PeersToSendRequest = 2

var log = logger.NewDefaultLogger()

type topicResolverSender struct {
	messenger   process.WireMessageHandler
	marshalizer marshal.Marshalizer
	topicName   string
	r           *rand.Rand
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(
	messenger process.WireMessageHandler,
	topicName string,
	marshalizer marshal.Marshalizer,
) (*topicResolverSender, error) {
	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	resolver := &topicResolverSender{
		messenger:   messenger,
		topicName:   topicName,
		marshalizer: marshalizer,
		r:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return resolver, nil
}

// SendOnRequestTopic is used to send request data over channels (topics) to other peers
// This method only sends the request, the received data should be handled by interceptors
func (trs *topicResolverSender) SendOnRequestTopic(rd *process.RequestData) error {
	buff, err := trs.marshalizer.Marshal(rd)
	if err != nil {
		return err
	}

	peersToSend := selectRandomPeers(trs.messenger.ConnectedPeers(), PeersToSendRequest, trs.r)
	if len(peersToSend) == 0 {
		return process.ErrNoConnectedPeerToSendRequest
	}

	for _, peer := range peersToSend {
		err = trs.messenger.SendToConnectedPeer(trs.topicName+requestTopicSuffix, buff, peer)
		if err != nil {
			log.Debug(err.Error())
		}
	}

	return nil
}

// Send is used to send an array buffer to a connected peer
// It is used when replying to a request
func (trs *topicResolverSender) Send(buff []byte, peer p2p.PeerID) error {
	return trs.messenger.SendToConnectedPeer(trs.topicName, buff, peer)
}

// RequestTopicSuffix returns the suffix that will be added to create a new channel for requests
func (trs *topicResolverSender) RequestTopicSuffix() string {
	return requestTopicSuffix
}

func selectRandomPeers(connectedPeers []p2p.PeerID, peersToSend int, randomizer process.IntRandomizer) []p2p.PeerID {
	selectedPeers := make([]p2p.PeerID, 0)

	if len(connectedPeers) == 0 {
		return selectedPeers
	}

	if len(connectedPeers) <= peersToSend {
		return connectedPeers
	}

	uniqueIndexes := make(map[int]struct{})
	//generating peersToSend number of unique indexes
	for len(uniqueIndexes) < peersToSend {
		newIndex := randomizer.Intn(len(connectedPeers))
		uniqueIndexes[newIndex] = struct{}{}
	}

	for index := range uniqueIndexes {
		selectedPeers = append(selectedPeers, connectedPeers[index])
	}

	return selectedPeers
}
