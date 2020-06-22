package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// HeaderResolverMock -
type HeaderResolverMock struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64, epoch uint32) error
	RequestDataFromEpochCalled   func(identifier []byte) error
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
}

// RequestDataFromEpoch -
func (hrm *HeaderResolverMock) RequestDataFromEpoch(identifier []byte) error {
	if hrm.RequestDataFromEpochCalled != nil {
		return hrm.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

// SetEpochHandler -
func (hrm *HeaderResolverMock) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrm.SetEpochHandlerCalled != nil {
		return hrm.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

// RequestDataFromHash -
func (hrm *HeaderResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

// ProcessReceivedMessage -
func (hrm *HeaderResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

// RequestDataFromNonce -
func (hrm *HeaderResolverMock) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	return hrm.RequestDataFromNonceCalled(nonce, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *HeaderResolverMock) IsInterfaceNil() bool {
	return hrm == nil
}
