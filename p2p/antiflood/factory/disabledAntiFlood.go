package factory

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type disabledAntiFlood struct {
}

// ResetForTopic won't do anything
func (n *disabledAntiFlood) ResetForTopic(topic string) {
}

// SetMaxMessagesForTopic won't do anything
func (n *disabledAntiFlood) SetMaxMessagesForTopic(topic string, maxNum uint32) {
}

// CanProcessMessage will always return nil
func (n *disabledAntiFlood) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
	return nil
}

// CanProcessMessageOnTopic will always return nil
func (n *disabledAntiFlood) CanProcessMessageOnTopic(peer p2p.PeerID, topic string) error {
	return nil
}

// IsInterfaceNil return true if there is no value under the interface
func (n *disabledAntiFlood) IsInterfaceNil() bool {
	return n == nil
}
