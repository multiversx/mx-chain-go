package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled         func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopicCalled func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64) error
	ApplyConsensusSizeCalled        func(size int)
	SetDebuggerCalled               func(debugger process.AntifloodDebugger) error
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}
	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// CanProcessMessagesOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64) error {
	if p2pahs.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}
	return p2pahs.CanProcessMessagesOnTopicCalled(peer, topic, numMessages, totalSize)
}

// ApplyConsensusSize -
func (p2pahs *P2PAntifloodHandlerStub) ApplyConsensusSize(size int) {
	if p2pahs.ApplyConsensusSizeCalled != nil {
		p2pahs.ApplyConsensusSizeCalled(size)
	}
}

// SetDebugger -
func (p2pahs *P2PAntifloodHandlerStub) SetDebugger(debugger process.AntifloodDebugger) error {
	if p2pahs.SetDebuggerCalled != nil {
		return p2pahs.SetDebuggerCalled(debugger)
	}

	return nil
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
