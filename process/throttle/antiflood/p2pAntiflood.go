package antiflood

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/disabled"
)

const unidentifiedTopic = "unidentifier topic"

var log = logger.GetOrCreate("process/throttle/antiflood")
var _ process.P2PAntifloodHandler = (*p2pAntiflood)(nil)

type p2pAntiflood struct {
	blacklistHandler    process.PeerBlackListCacher
	floodPreventers     []process.FloodPreventer
	topicPreventer      process.TopicFloodPreventer
	mutDebugger         sync.RWMutex
	debugger            process.AntifloodDebugger
	peerValidatorMapper process.PeerValidatorMapper
	mapTopicsFromAll    map[string]struct{}
	mutTopicCheck       sync.RWMutex
}

// NewP2PAntiflood creates a new p2p anti flood protection mechanism built on top of a flood preventer implementation.
// It contains only the p2p anti flood logic that should be applied
func NewP2PAntiflood(
	blacklistHandler process.PeerBlackListCacher,
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
		return nil, process.ErrNilBlackListCacher
	}

	return &p2pAntiflood{
		blacklistHandler:    blacklistHandler,
		floodPreventers:     floodPreventers,
		topicPreventer:      topicFloodPreventer,
		debugger:            &disabled.AntifloodDebugger{},
		mapTopicsFromAll:    make(map[string]struct{}),
		peerValidatorMapper: &disabled.PeerValidatorMapper{},
	}, nil
}

// CanProcessMessage signals if a p2p message can be processed or not
func (af *p2pAntiflood) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
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
		af.recordDebugEvent(
			fromConnectedPeer,
			message.Topics(),
			1,
			uint64(len(message.Data())),
			message.SeqNo(),
			af.blacklistHandler.Has(fromConnectedPeer),
		)

		return lastErrFound
	}

	originatorIsBlacklisted := af.blacklistHandler.Has(message.Peer())
	if originatorIsBlacklisted {
		af.recordDebugEvent(message.Peer(), message.Topics(), 1, uint64(len(message.Data())), message.SeqNo(), true)
		return fmt.Errorf("%w for pid %s", process.ErrOriginatorIsBlacklisted, message.Peer().Pretty())
	}

	return nil
}

// IsOriginatorEligibleForTopic returns error if pid is not allowed to send messages on topic
func (af *p2pAntiflood) IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error {
	af.mutTopicCheck.RLock()
	defer af.mutTopicCheck.RUnlock()

	_, ok := af.mapTopicsFromAll[topic]
	if ok {
		return nil
	}

	peerInfo := af.peerValidatorMapper.GetPeerInfo(pid)
	if peerInfo.PeerType == core.ValidatorPeer {
		return nil
	}

	return process.ErrOnlyValidatorsCanUseThisTopic
}

// SetTopicsForAll sets the topics which are enabled for all
func (af *p2pAntiflood) SetTopicsForAll(topics ...string) {
	af.mutTopicCheck.Lock()
	defer af.mutTopicCheck.Unlock()

	for _, topic := range topics {
		af.mapTopicsFromAll[topic] = struct{}{}
	}
}

// SetPeerValidatorMapper sets the peer validator mapper
func (af *p2pAntiflood) SetPeerValidatorMapper(validatorMapper process.PeerValidatorMapper) error {
	if check.IfNil(validatorMapper) {
		return process.ErrNilPeerValidatorMapper
	}

	af.mutTopicCheck.Lock()
	defer af.mutTopicCheck.Unlock()

	af.peerValidatorMapper = validatorMapper
	return nil
}

func (af *p2pAntiflood) recordDebugEvent(pid core.PeerID, topics []string, numRejected uint32, sizeRejected uint64, sequence []byte, isBlacklisted bool) {
	if len(topics) == 0 {
		topics = []string{unidentifiedTopic}
	}

	af.mutDebugger.RLock()
	defer af.mutDebugger.RUnlock()

	af.debugger.AddData(pid, topics[0], numRejected, sizeRejected, sequence, isBlacklisted)
}

func (af *p2pAntiflood) canProcessMessage(fp process.FloodPreventer, message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	//protect from directly connected peer
	err := fp.IncreaseLoad(fromConnectedPeer, uint64(len(message.Data())))
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
		err = fp.IncreaseLoad(message.Peer(), uint64(len(message.Data())))
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
func (af *p2pAntiflood) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
	err := af.topicPreventer.IncreaseLoad(peer, topic, numMessages)
	if err != nil {
		log.Trace("topicFloodPreventer.Accumulate peer",
			"error", err,
			"pid", p2p.PeerIdToShortString(peer),
			"topic", topic,
		)

		af.recordDebugEvent(peer, []string{topic}, numMessages, totalSize, sequence, af.blacklistHandler.Has(peer))

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

// SetDebugger sets the antiflood debugger
func (af *p2pAntiflood) SetDebugger(debugger process.AntifloodDebugger) error {
	if check.IfNil(debugger) {
		return process.ErrNilDebugger
	}

	af.mutDebugger.Lock()
	log.LogIfError(af.debugger.Close())
	af.debugger = debugger
	af.mutDebugger.Unlock()

	return nil
}

// BlacklistPeer will add a peer to the black list
func (af *p2pAntiflood) BlacklistPeer(peer core.PeerID, reason string, duration time.Duration) {
	peerIsBlacklisted := af.blacklistHandler.Has(peer)

	err := af.blacklistHandler.Upsert(peer, duration)
	if err != nil {
		log.Warn("error adding in blacklist",
			"pid", peer.Pretty(),
			"time", duration,
			"reason", reason,
			"error", "err",
		)
		return
	}

	if !peerIsBlacklisted {
		log.Debug("blacklisted peer",
			"pid", peer.Pretty(),
			"time", duration,
			"reason", reason,
		)
	}
}

// Close will call the close function on all sub components
func (af *p2pAntiflood) Close() error {
	return af.debugger.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *p2pAntiflood) IsInterfaceNil() bool {
	return af == nil
}
