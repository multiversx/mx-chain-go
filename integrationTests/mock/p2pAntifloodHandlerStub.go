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
	ApplyConsensusSizeCalled           func(size int)
	SetDebuggerCalled                  func(debugger process.AntifloodDebugger) error
	BlacklistPeerCalled                func(peer core.PeerID, reason string, duration time.Duration)
	IsOriginatorEligibleForTopicCalled func(pid core.PeerID, topic string) error
	SetPeerValidatorMapperCalled       func(validatorMapper process.PeerValidatorMapper) error
	SetConsensusSizeNotifierCalled     func(chainParametersNotifier process.ChainParametersSubscriber, shardID uint32)
}

// CanProcessMessage -
func (stub *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if stub.CanProcessMessageCalled == nil {
		return nil
	}
	return stub.CanProcessMessageCalled(message, fromConnectedPeer)
}

// IsOriginatorEligibleForTopic -
func (stub *P2PAntifloodHandlerStub) IsOriginatorEligibleForTopic(pid core.PeerID, topic string) error {
	if stub.IsOriginatorEligibleForTopicCalled != nil {
		return stub.IsOriginatorEligibleForTopicCalled(pid, topic)
	}
	return nil
}

// CanProcessMessagesOnTopic -
func (stub *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
	if stub.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}
	return stub.CanProcessMessagesOnTopicCalled(peer, topic, numMessages, totalSize, sequence)
}

// ApplyConsensusSize -
func (stub *P2PAntifloodHandlerStub) ApplyConsensusSize(size int) {
	if stub.ApplyConsensusSizeCalled != nil {
		stub.ApplyConsensusSizeCalled(size)
	}
}

// SetConsensusSizeNotifier -
func (p2pahs *P2PAntifloodHandlerStub) SetConsensusSizeNotifier(chainParametersNotifier process.ChainParametersSubscriber, shardID uint32) {
	if p2pahs.SetConsensusSizeNotifierCalled != nil {
		p2pahs.SetConsensusSizeNotifierCalled(chainParametersNotifier, shardID)
	}
}

// SetDebugger -
func (stub *P2PAntifloodHandlerStub) SetDebugger(debugger process.AntifloodDebugger) error {
	if stub.SetDebuggerCalled != nil {
		return stub.SetDebuggerCalled(debugger)
	}

	return nil
}

// BlacklistPeer -
func (stub *P2PAntifloodHandlerStub) BlacklistPeer(peer core.PeerID, reason string, duration time.Duration) {
	if stub.BlacklistPeerCalled != nil {
		stub.BlacklistPeerCalled(peer, reason, duration)
	}
}

// ResetForTopic -
func (stub *P2PAntifloodHandlerStub) ResetForTopic(_ string) {

}

// SetMaxMessagesForTopic -
func (stub *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(_ string, _ uint32) {

}

// SetPeerValidatorMapper -
func (stub *P2PAntifloodHandlerStub) SetPeerValidatorMapper(validatorMapper process.PeerValidatorMapper) error {
	if stub.SetPeerValidatorMapperCalled != nil {
		return stub.SetPeerValidatorMapperCalled(validatorMapper)
	}
	return nil
}

// SetTopicsForAll -
func (stub *P2PAntifloodHandlerStub) SetTopicsForAll(_ ...string) {

}

// Close -
func (stub *P2PAntifloodHandlerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (stub *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
