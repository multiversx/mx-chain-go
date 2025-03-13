package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/pkg/errors"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

var errNotImplemented = errors.New("not implemented")

// HeaderResolverStub -
type HeaderResolverStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) ([]byte, error)
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
	SetDebugHandlerCalled        func(handler dataRetriever.DebugHandler) error
	CloseCalled                  func() error
}

// SetEpochHandler -
func (hrs *HeaderResolverStub) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrs.SetEpochHandlerCalled != nil {
		return hrs.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

// ProcessReceivedMessage -
func (hrs *HeaderResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) ([]byte, error) {
	if hrs.ProcessReceivedMessageCalled != nil {
		return hrs.ProcessReceivedMessageCalled(message)
	}

	return nil, errNotImplemented
}

// SetDebugHandler -
func (hrs *HeaderResolverStub) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if hrs.SetDebugHandlerCalled != nil {
		return hrs.SetDebugHandlerCalled(handler)
	}

	return nil
}

func (hrs *HeaderResolverStub) Close() error {
	if hrs.CloseCalled != nil {
		return hrs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrs *HeaderResolverStub) IsInterfaceNil() bool {
	return hrs == nil
}
