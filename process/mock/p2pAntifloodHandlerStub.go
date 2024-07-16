package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled            func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopicCalled    func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	SetConsensusSizeNotifierCalled     func(subscriber process.ChainParametersSubscriber, shardID uint32)
	SetDebuggerCalled                  func(debugger process.AntifloodDebugger) error
	BlacklistPeerCalled                func(peer core.PeerID, reason string, duration time.Duration)
	IsOriginatorEligibleForTopicCalled func(pid core.PeerID, topic string) error
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}
	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// IsOriginatorEligibleForTopic -
func (p2pahs *P2PAntifloodHandlerStub) IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error {
	if p2pahs.IsOriginatorEligibleForTopicCalled != nil {
		return p2pahs.IsOriginatorEligibleForTopicCalled(pid, topic)
	}
	return nil
}

// CanProcessMessagesOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
	if p2pahs.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}
	return p2pahs.CanProcessMessagesOnTopicCalled(peer, topic, numMessages, totalSize, sequence)
}

// SetConsensusSizeNotifier -
func (p2pahs *P2PAntifloodHandlerStub) SetConsensusSizeNotifier(subscriber process.ChainParametersSubscriber, shardID uint32) {
	if p2pahs.SetConsensusSizeNotifierCalled != nil {
		p2pahs.SetConsensusSizeNotifierCalled(subscriber, shardID)
	}
}

// SetDebugger -
func (p2pahs *P2PAntifloodHandlerStub) SetDebugger(debugger process.AntifloodDebugger) error {
	if p2pahs.SetDebuggerCalled != nil {
		return p2pahs.SetDebuggerCalled(debugger)
	}

	return nil
}

// BlacklistPeer -
func (p2pahs *P2PAntifloodHandlerStub) BlacklistPeer(peer core.PeerID, reason string, duration time.Duration) {
	if p2pahs.BlacklistPeerCalled != nil {
		p2pahs.BlacklistPeerCalled(peer, reason, duration)
	}
}

// Close -
func (p2pahs *P2PAntifloodHandlerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
