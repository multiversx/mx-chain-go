package disabled

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var _ consensus.P2PAntifloodHandler = (*AntiFlood)(nil)

// AntiFlood is a mock implementation of the antiflood interface
type AntiFlood struct {
}

// ResetForTopic won't do anything
func (af *AntiFlood) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic won't do anything
func (af *AntiFlood) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage will always return nil
func (af *AntiFlood) CanProcessMessage(_ p2p.MessageP2P, _ p2p.PeerID) error {
	return nil
}

// CanProcessMessagesOnTopic will always return nil
func (af *AntiFlood) CanProcessMessagesOnTopic(_ p2p.PeerID, _ string, _ uint32) error {
	return nil
}

// ApplyConsensusSize does nothing
func (af *AntiFlood) ApplyConsensusSize(_ int) {
}

// IsInterfaceNil return true if there is no value under the interface
func (af *AntiFlood) IsInterfaceNil() bool {
	return af == nil
}
