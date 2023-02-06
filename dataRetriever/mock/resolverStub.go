package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

// ResolverStub -
type ResolverStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	SetDebugHandlerCalled        func(handler dataRetriever.DebugHandler) error
	CloseCalled                  func() error
}

// ProcessReceivedMessage -
func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return rs.ProcessReceivedMessageCalled(message)
}

// SetDebugHandler -
func (rs *ResolverStub) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if rs.SetDebugHandlerCalled != nil {
		return rs.SetDebugHandlerCalled(handler)
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
