package factory

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type disabledAntiFlood struct {
}

// ResetForTopic won't do anything
func (daf *disabledAntiFlood) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic won't do anything
func (daf *disabledAntiFlood) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage will always return nil
func (daf *disabledAntiFlood) CanProcessMessage(_ p2p.MessageP2P, _ p2p.PeerID) error {
	return nil
}

// CanProcessMessagesOnTopic will always return nil
func (daf *disabledAntiFlood) CanProcessMessagesOnTopic(_ p2p.PeerID, _ string, _ uint32) error {
	return nil
}

// IsInterfaceNil return true if there is no value under the interface
func (daf *disabledAntiFlood) IsInterfaceNil() bool {
	return daf == nil
}
