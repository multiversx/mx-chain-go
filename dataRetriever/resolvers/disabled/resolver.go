package disabled

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type resolver struct {
}

// NewDisabledResolver returns a new instance of disabled resolver
func NewDisabledResolver() *resolver {
	return &resolver{}
}

// ProcessReceivedMessage returns nil as it is disabled
func (r *resolver) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// SetResolverDebugHandler returns nil as it is disabled
func (r *resolver) SetResolverDebugHandler(_ dataRetriever.ResolverDebugHandler) error {
	return nil
}

// Close returns nil as it is disabled
func (r *resolver) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *resolver) IsInterfaceNil() bool {
	return r == nil
}
