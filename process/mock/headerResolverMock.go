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

func (hrs *HeaderResolverMock) RequestDataFromEpoch(identifier []byte) error {
	if hrs.RequestDataFromEpochCalled != nil {
		return hrs.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

func (hrs *HeaderResolverMock) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrs.SetEpochHandlerCalled != nil {
		return hrs.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

func (hrm *HeaderResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

func (hrm *HeaderResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *HeaderResolverMock) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	return hrm.RequestDataFromNonceCalled(nonce, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *HeaderResolverMock) IsInterfaceNil() bool {
	return hrm == nil
}
