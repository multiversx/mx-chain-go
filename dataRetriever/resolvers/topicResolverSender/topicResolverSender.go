package topicResolverSender

import (
	"bytes"

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
	messenger             dataRetriever.MessageHandler
	marshalizer           marshal.Marshalizer
	topicName             string
	excludePeersFromTopic string
	randomizer            dataRetriever.IntRandomizer
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(
	messenger dataRetriever.MessageHandler,
	topicName string,
	excludePeersFromTopic string,
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
		messenger:             messenger,
		topicName:             topicName,
		excludePeersFromTopic: excludePeersFromTopic,
		marshalizer:           marshalizer,
		randomizer:            randomizer,
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
	excludedConnectedPeers := make([]p2p.PeerID, 0)

	if len(trs.excludePeersFromTopic) > 0 {
		excludedConnectedPeers = trs.messenger.ConnectedPeersOnTopic(excludedTopicSendRequest)
	}

	peersToSend, err := selectRandomPeers(allConnectedPeers, excludedConnectedPeers, NumPeersToQuery, trs.randomizer)
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

func selectRandomPeers(
	allConnectedPeers []p2p.PeerID,
	excludedConnectedPeers []p2p.PeerID,
	peersToSend int,
	randomizer dataRetriever.IntRandomizer,
) ([]p2p.PeerID, error) {

	selectedPeers := make([]p2p.PeerID, 0)

	if len(allConnectedPeers) == 0 {
		return selectedPeers, nil
	}

	diffPeers := makeDiffList(allConnectedPeers, excludedConnectedPeers)
	if len(diffPeers) == 0 {
		//if no peer "from other" shard is found, this fallback ensures
		//that the requests will be send to some of the connected peers
		//hoping that they might have the required info
		diffPeers = allConnectedPeers
	}
	if len(diffPeers) <= peersToSend {
		return diffPeers, nil
	}

	uniqueIndexes := make(map[int]struct{})
	//generating peersToSend number of unique indexes
	for len(uniqueIndexes) < peersToSend {
		newIndex, err := randomizer.Intn(len(diffPeers))
		if err != nil {
			return nil, err
		}

		uniqueIndexes[newIndex] = struct{}{}
	}

	for index := range uniqueIndexes {
		selectedPeers = append(selectedPeers, diffPeers[index])
	}

	return selectedPeers, nil
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
