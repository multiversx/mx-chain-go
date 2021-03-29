package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
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
func (af *AntiFlood) CanProcessMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// IsOriginatorEligibleForTopic will always return nil
func (af *AntiFlood) IsOriginatorEligibleForTopic(_ core.PeerID, _ string) error {
	return nil
}

// SetTopicsForAll does nothing
func (af *AntiFlood) SetTopicsForAll(_ ...string) {
}

// SetPeerValidatorMapper does nothing
func (af *AntiFlood) SetPeerValidatorMapper(_ process.PeerValidatorMapper) error {
	return nil
}

// CanProcessMessagesOnTopic will always return nil
func (af *AntiFlood) CanProcessMessagesOnTopic(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
	return nil
}

// ApplyConsensusSize does nothing
func (af *AntiFlood) ApplyConsensusSize(_ int) {
}

// SetDebugger returns nil
func (af *AntiFlood) SetDebugger(_ process.AntifloodDebugger) error {
	return nil
}

// BlacklistPeer does nothing
func (af *AntiFlood) BlacklistPeer(_ core.PeerID, _ string, _ time.Duration) {
}

// Close does nothing
func (af *AntiFlood) Close() error {
	return nil
}

// IsInterfaceNil return true if there is no value under the interface
func (af *AntiFlood) IsInterfaceNil() bool {
	return af == nil
}
