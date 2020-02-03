package factory

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type disabledAntiFlood struct {
}

// ResetForTopic won't do anything
func (n *disabledAntiFlood) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic won't do anything
func (n *disabledAntiFlood) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage will always return nil
func (n *disabledAntiFlood) CanProcessMessage(_ p2p.MessageP2P, _ p2p.PeerID) error {
	return nil
}

// CanProcessMessageOnTopic will always return nil
func (n *disabledAntiFlood) CanProcessMessageOnTopic(_ p2p.PeerID, _ string) error {
	return nil
}

// IsInterfaceNil return true if there is no value under the interface
func (n *disabledAntiFlood) IsInterfaceNil() bool {
	return n == nil
}
