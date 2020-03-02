package antiflood

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var log = logger.GetOrCreate("process/throttle/antiflood")

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
		log.Trace("floodPreventer.AccumulateGlobal connected peer",
			"error", p2p.ErrSystemBusy,
			"pid", p2p.PeerIdToShortString(fromConnectedPeer),
			"message payload bytes", uint64(len(message.Data())),
		)
		return fmt.Errorf("%w in p2pAntiflood for connected peer %s",
			p2p.ErrSystemBusy,
			p2p.PeerIdToShortString(fromConnectedPeer),
		)
	}

	if fromConnectedPeer != message.Peer() {
		//protect from the flooding messages that originate from the same source but come from different peers
		ok = floodPreventer.Accumulate(message.Peer().Pretty(), uint64(len(message.Data())))
		if !ok {
			log.Trace("floodPreventer.AccumulateGlobal originator",
				"error", p2p.ErrSystemBusy,
				"pid", p2p.MessageOriginatorPid(message),
				"message payload bytes", uint64(len(message.Data())),
			)
			return fmt.Errorf("%w in p2pAntiflood for originator %s",
				p2p.ErrSystemBusy,
				p2p.MessageOriginatorPid(message),
			)
		}
	}

	return nil
}

// CanProcessMessageOnTopic signals if a p2p message can be processed or not for a given topic
func (af *p2pAntiflood) CanProcessMessageOnTopic(peer p2p.PeerID, topic string) error {
	topicFloodPreventer := af.TopicFloodPreventer
	if check.IfNil(topicFloodPreventer) {
		return p2p.ErrNilTopicFloodPreventer
	}

	ok := topicFloodPreventer.Accumulate(peer.Pretty(), topic)
	if !ok {
		log.Trace("topicFloodPreventer.Accumulate peer",
			"error", p2p.ErrSystemBusy,
			"pid", p2p.PeerIdToShortString(peer),
			"topic", topic,
		)
		return fmt.Errorf("%w in p2pAntiflood for connected peer %s",
			p2p.ErrSystemBusy,
			p2p.PeerIdToShortString(peer),
		)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *p2pAntiflood) IsInterfaceNil() bool {
	return af == nil || check.IfNil(af.FloodPreventer) || check.IfNil(af.TopicFloodPreventer)
}
