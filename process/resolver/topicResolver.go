package resolver

import (
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// RequestTopicSuffix represents the topic name suffix
const RequestTopicSuffix = "_REQUEST"

// PeersToSendRequest number of peers to send the message
const PeersToSendRequest = 2

var log = logger.NewDefaultLogger()

// topicResolver is a struct coupled with a p2p.Topic that can process requests
type topicResolver struct {
	messenger   p2p.Messenger
	name        string
	marshalizer marshal.Marshalizer
	r           *rand.Rand

	resolveRequest func(rd process.RequestData) ([]byte, error)
}

// NewTopicResolver returns a new topic resolver instance
func NewTopicResolver(
	name string,
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
) (*topicResolver, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	resolver := &topicResolver{
		name:        name,
		messenger:   messenger,
		marshalizer: marshalizer,
		r:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	err := resolver.registerRequestValidator()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (tr *topicResolver) registerRequestValidator() error {
	if tr.messenger.HasTopicValidator(tr.name + RequestTopicSuffix) {
		return process.ErrValidatorAlreadySet
	}

	if !tr.messenger.HasTopic(tr.name + RequestTopicSuffix) {
		err := tr.messenger.CreateTopic(tr.name+RequestTopicSuffix, false)
		if err != nil {
			return err
		}
	}

	return tr.messenger.SetTopicValidator(tr.name+RequestTopicSuffix, tr.requestValidator)
}

func (tr *topicResolver) requestValidator(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	rd := process.RequestData{}
	err := tr.marshalizer.Unmarshal(&rd, message.Data())
	if err != nil {
		return err
	}

	peerSendBack := message.Peer()

	if tr.resolveRequest == nil {
		return process.ErrNilResolverHandler
	}

	buff, err := tr.resolveRequest(rd)
	if err != nil {
		return err
	}

	return tr.messenger.SendToConnectedPeer(tr.name, buff, peerSendBack)
}

// RequestData is used to request data over channels (topics) from other peers
// This method only sends the request, the received data should be handled by interceptors
func (tr *topicResolver) RequestData(rd process.RequestData) error {
	buff, err := tr.marshalizer.Marshal(&rd)
	if err != nil {
		return err
	}

	peersToSend := selectRandomPeers(tr.messenger.ConnectedPeers(), PeersToSendRequest, tr.r)
	if len(peersToSend) == 0 {
		return process.ErrNoConnectedPeerToSendRequest
	}

	for _, peer := range peersToSend {
		err = tr.messenger.SendToConnectedPeer(tr.name+RequestTopicSuffix, buff, peer)
		if err != nil {
			log.Debug(err.Error())
		}
	}

	return nil
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

// SetResolverHandler sets the handler that will be called when a new request comes from other peers to
// current node
func (tr *topicResolver) SetResolverHandler(handler func(rd process.RequestData) ([]byte, error)) {
	tr.resolveRequest = handler
}

// ResolverHandler gets the handler that will be called when a new request comes from other peers to
// current node
func (tr *topicResolver) ResolverHandler() func(rd process.RequestData) ([]byte, error) {
	return tr.resolveRequest
}
