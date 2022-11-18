package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// ResolverStub -
type ResolverStub struct {
	ProcessReceivedMessageCalled  func(message p2p.MessageP2P) error
	SetResolverDebugHandlerCalled func(handler dataRetriever.ResolverDebugHandler) error
	CloseCalled                   func() error
}

// ProcessReceivedMessage -
func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return rs.ProcessReceivedMessageCalled(message)
}

// SetResolverDebugHandler -
func (rs *ResolverStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if rs.SetResolverDebugHandlerCalled != nil {
		return rs.SetResolverDebugHandlerCalled(handler)
	}

	return nil
}

// Close -
func (rs *ResolverStub) Close() error {
	if rs.CloseCalled != nil {
		return rs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ResolverStub) IsInterfaceNil() bool {
	return rs == nil
}
