package antiflood

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("process/throttle/antiflood")

var _ process.P2PAntifloodHandler = (*p2pAntiflood)(nil)

type p2pAntiflood struct {
	blacklistHandler process.BlackListHandler
	floodPreventers  []process.FloodPreventer
	topicPreventer   process.TopicFloodPreventer
}

// NewP2PAntiflood creates a new p2p anti flood protection mechanism built on top of a flood preventer implementation.
// It contains only the p2p anti flood logic that should be applied
func NewP2PAntiflood(
	blacklistHandler process.BlackListHandler,
	topicFloodPreventer process.TopicFloodPreventer,
	floodPreventers ...process.FloodPreventer,
) (*p2pAntiflood, error) {

	if len(floodPreventers) == 0 {
		return nil, process.ErrEmptyFloodPreventerList
	}
	if check.IfNil(topicFloodPreventer) {
		return nil, process.ErrNilTopicFloodPreventer
	}
	if check.IfNil(blacklistHandler) {
		return nil, process.ErrNilBlackListHandler
	}

	return &p2pAntiflood{
		blacklistHandler: blacklistHandler,
		floodPreventers:  floodPreventers,
		topicPreventer:   topicFloodPreventer,
	}, nil
}

// CanProcessMessage signals if a p2p message can be processed or not
func (af *p2pAntiflood) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	if message == nil {
		return p2p.ErrNilMessage
	}

	var lastErrFound error
	for _, fp := range af.floodPreventers {
		err := af.canProcessMessage(fp, message, fromConnectedPeer)
		if err != nil {
			lastErrFound = err
		}
	}

	if lastErrFound != nil {
		return lastErrFound
	}

	originatorIsBlacklisted := af.blacklistHandler.Has(message.Peer().Pretty())
	if originatorIsBlacklisted {
		return fmt.Errorf("%w for pid %s", process.ErrOriginatorIsBlacklisted, message.Peer())
	}

	return nil
}

func (af *p2pAntiflood) canProcessMessage(fp process.FloodPreventer, message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	//protect from directly connected peer
	err := fp.IncreaseLoad(fromConnectedPeer.Pretty(), uint64(len(message.Data())))
	if err != nil {
		log.Trace("floodPreventer.IncreaseLoad connected peer",
			"error", err,
			"pid", p2p.PeerIdToShortString(fromConnectedPeer),
			"message payload bytes", uint64(len(message.Data())),
		)
		return fmt.Errorf("%w in p2pAntiflood for connected peer %s",
			err,
			p2p.PeerIdToShortString(fromConnectedPeer),
		)
	}

	if fromConnectedPeer != message.Peer() {
		//protect from the flooding messages that originate from the same source but come from different peers
		err = fp.IncreaseLoad(message.Peer().Pretty(), uint64(len(message.Data())))
		if err != nil {
			log.Trace("floodPreventer.IncreaseLoad originator",
				"error", err,
				"pid", p2p.MessageOriginatorPid(message),
				"message payload bytes", uint64(len(message.Data())),
			)
			return fmt.Errorf("%w in p2pAntiflood for originator %s",
				err,
				p2p.MessageOriginatorPid(message),
			)
		}
	}

	return nil
}

// CanProcessMessagesOnTopic signals if a p2p message can be processed or not for a given topic
func (af *p2pAntiflood) CanProcessMessagesOnTopic(peer p2p.PeerID, topic string, numMessages uint32) error {
	err := af.topicPreventer.IncreaseLoad(peer.Pretty(), topic, numMessages)
	if err != nil {
		log.Trace("topicFloodPreventer.Accumulate peer",
			"error", err,
			"pid", p2p.PeerIdToShortString(peer),
			"topic", topic,
		)
		return fmt.Errorf("%w in p2pAntiflood for connected peer %s",
			err,
			p2p.PeerIdToShortString(peer),
		)
	}

	return nil
}

// SetMaxMessagesForTopic will update the maximum number of messages that can be received from a peer in a topic
func (af *p2pAntiflood) SetMaxMessagesForTopic(topic string, numMessages uint32) {
	af.topicPreventer.SetMaxMessagesForTopic(topic, numMessages)
}

// ResetForTopic clears all map values for a given topic
func (af *p2pAntiflood) ResetForTopic(topic string) {
	af.topicPreventer.ResetForTopic(topic)
}

// ApplyConsensusSize applies the consensus size on all contained flood preventers
func (af *p2pAntiflood) ApplyConsensusSize(size int) {
	for _, fp := range af.floodPreventers {
		fp.ApplyConsensusSize(size)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *p2pAntiflood) IsInterfaceNil() bool {
	return af == nil
}
