package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type HeaderResolverMock struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64, epoch uint32) error
	RequestDataFromEpochCalled   func(identifier []byte) error
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
}

func (hrm *HeaderResolverMock) RequestDataFromEpoch(identifier []byte) error {
	if hrm.RequestDataFromEpochCalled != nil {
		return hrm.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

func (hrm *HeaderResolverMock) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrm.SetEpochHandlerCalled != nil {
		return hrm.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

func (hrm *HeaderResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

func (hrm *HeaderResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *HeaderResolverMock) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	return hrm.RequestDataFromNonceCalled(nonce, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *HeaderResolverMock) IsInterfaceNil() bool {
	return hrm == nil
}
