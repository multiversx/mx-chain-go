package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ dataRetriever.P2PAntifloodHandler = (*antiFloodHandler)(nil)

type antiFloodHandler struct {
}

// NewAntiFloodHandler returns a new instance of antiFloodHandler
func NewAntiFloodHandler() *antiFloodHandler {
	return &antiFloodHandler{}
}

// CanProcessMessage returns nil regardless of the input
func (a *antiFloodHandler) CanProcessMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// CanProcessMessagesOnTopic returns nil regardless of the input
func (a *antiFloodHandler) CanProcessMessagesOnTopic(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
	return nil
}

// ApplyConsensusSize does nothing
func (a *antiFloodHandler) ApplyConsensusSize(_ int) {
}

// SetDebugger returns nil
func (a *antiFloodHandler) SetDebugger(_ process.AntifloodDebugger) error {
	return nil
}

// BlacklistPeer does nothing
func (a *antiFloodHandler) BlacklistPeer(_ core.PeerID, _ string, _ time.Duration) {
}

// IsOriginatorEligibleForTopic returns nil
func (a *antiFloodHandler) IsOriginatorEligibleForTopic(_ core.PeerID, _ string) error {
	return nil
}

// Close returns nil
func (af *antiFloodHandler) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *antiFloodHandler) IsInterfaceNil() bool {
	return a == nil
}
