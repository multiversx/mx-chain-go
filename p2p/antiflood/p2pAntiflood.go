package antiflood

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type p2pAntiflood struct {
	p2p.FloodPreventer
	p2p.TopicFloodPreventer
}

// NewP2PAntiflood creates a new p2p anti flood protection mechanism built on top of a flood preventer implementation.
// It contains only the p2p anti flood logic that should be applied
func NewP2PAntiflood(floodPreventer p2p.FloodPreventer, topicFloodPreventer p2p.TopicFloodPreventer) (*p2pAntiflood, error) {
	if check.IfNil(floodPreventer) {
		return nil, p2p.ErrNilFloodPreventer
	}
	if check.IfNil(topicFloodPreventer) {
		return nil, p2p.ErrNilTopicFloodPreventer
	}

	return &p2pAntiflood{
		FloodPreventer:      floodPreventer,
		TopicFloodPreventer: topicFloodPreventer,
	}, nil
}

// CanProcessMessage signals if a p2p message can be processed or not
func (af *p2pAntiflood) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	floodPreventer := af.FloodPreventer
	if check.IfNil(floodPreventer) {
		return p2p.ErrNilFloodPreventer
	}
	if message == nil {
		return p2p.ErrNilMessage
	}

	//protect from directly connected peer
	ok := floodPreventer.AccumulateGlobal(fromConnectedPeer.Pretty(), uint64(len(message.Data())))
	if !ok {
		return fmt.Errorf("%w in p2pAntiflood for connected peer", p2p.ErrSystemBusy)
	}

	if fromConnectedPeer != message.Peer() {
		//protect from the flooding messages that originate from the same source but come from different peers
		ok = floodPreventer.Accumulate(message.Peer().Pretty(), uint64(len(message.Data())))
		if !ok {
			return fmt.Errorf("%w in p2pAntiflood for originator", p2p.ErrSystemBusy)
		}
	}

	return nil
}

// CanProcessMessageOnTopic signals if a p2p message can be processed or not for a given topic
func (af *p2pAntiflood) CanProcessMessageOnTopic(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID, topic string) error {
	topicFloodPreventer := af.TopicFloodPreventer
	if check.IfNil(topicFloodPreventer) {
		return p2p.ErrNilTopicFloodPreventer
	}
	if message == nil {
		return p2p.ErrNilMessage
	}

	peerId := message.Peer().Pretty()
	ok := topicFloodPreventer.Accumulate(peerId, topic)
	if !ok {
		return fmt.Errorf("%w in p2pAntiflood for originator. peer id = %s", p2p.ErrSystemBusy, peerId)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *p2pAntiflood) IsInterfaceNil() bool {
	return af == nil || check.IfNil(af.FloodPreventer)
}
