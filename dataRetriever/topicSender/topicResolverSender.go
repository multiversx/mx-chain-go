package topicsender

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

var _ dataRetriever.TopicResolverSender = (*topicResolverSender)(nil)

// ArgTopicResolverSender is the argument structure used to create new TopicResolverSender instance
type ArgTopicResolverSender struct {
	ArgBaseTopicSender
}

type topicResolverSender struct {
	*baseTopicSender
}

// NewTopicResolverSender returns a new topic resolver instance
func NewTopicResolverSender(arg ArgTopicResolverSender) (*topicResolverSender, error) {
	err := checkBaseTopicSenderArgs(arg.ArgBaseTopicSender)
	if err != nil {
		return nil, err
	}

	return &topicResolverSender{
		baseTopicSender: createBaseTopicSender(arg.ArgBaseTopicSender),
	}, nil
}

// Send is used to send an array buffer to a connected peer
// It is used when replying to a request
func (trs *topicResolverSender) Send(buff []byte, peer core.PeerID) error {
	if common.ShouldNotRespondToRequests.IsSet() {
		log.Debug("testing- not responding to requests anymore", "self", trs.messenger.ID().Pretty(), "peer", peer.Pretty(), "topic", trs.topicName)
		return fmt.Errorf("testing- forced to not respond to requests anymore")
	}

	return trs.sendToConnectedPeer(trs.topicName, buff, peer)
}

// IsInterfaceNil returns true if there is no value under the interface
func (trs *topicResolverSender) IsInterfaceNil() bool {
	return trs == nil
}
