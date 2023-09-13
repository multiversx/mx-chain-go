package disabled

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type antiFlooder struct {
}

// NewAntiFlooder creates a new instance of disabled antiflooder
func NewAntiFlooder() *antiFlooder {
	return &antiFlooder{}
}

// CanProcessMessage returns nil
func (a *antiFlooder) CanProcessMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// IsOriginatorEligibleForTopic does nothing and returns nil
func (a *antiFlooder) IsOriginatorEligibleForTopic(_ core.PeerID, _ string) error {
	return nil
}

// CanProcessMessagesOnTopic does nothing and returns nil
func (a *antiFlooder) CanProcessMessagesOnTopic(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
	return nil
}

// ApplyConsensusSize does nothing
func (a *antiFlooder) ApplyConsensusSize(_ int) {
}

// SetDebugger does nothing and returns nil
func (a *antiFlooder) SetDebugger(_ process.AntifloodDebugger) error {
	return nil
}

// BlacklistPeer does nothing
func (a *antiFlooder) BlacklistPeer(_ core.PeerID, _ string, _ time.Duration) {
}

// ResetForTopic does nothing
func (a *antiFlooder) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic does nothing
func (a *antiFlooder) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// SetPeerValidatorMapper does nothing and returns nil
func (a *antiFlooder) SetPeerValidatorMapper(_ process.PeerValidatorMapper) error {
	return nil
}

// SetTopicsForAll does nothing
func (a *antiFlooder) SetTopicsForAll(_ ...string) {
}

// Close does nothing and returns nil
func (a *antiFlooder) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *antiFlooder) IsInterfaceNil() bool {
	return a == nil
}
