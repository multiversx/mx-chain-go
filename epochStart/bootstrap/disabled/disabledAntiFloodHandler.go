package disabled

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var _ dataRetriever.P2PAntifloodHandler = (*antiFloodHandler)(nil)

type antiFloodHandler struct {
}

// NewAntiFloodHandler returns a new instance of antiFloodHandler
func NewAntiFloodHandler() *antiFloodHandler {
	return &antiFloodHandler{}
}

// CanProcessMessage returns nil regardless of the input
func (a *antiFloodHandler) CanProcessMessage(_ p2p.MessageP2P, _ p2p.PeerID) error {
	return nil
}

// CanProcessMessagesOnTopic returns nil regardless of the input
func (a *antiFloodHandler) CanProcessMessagesOnTopic(_ p2p.PeerID, _ string, _ uint32) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *antiFloodHandler) IsInterfaceNil() bool {
	return a == nil
}
