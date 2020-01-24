package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type HeaderResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64, epoch uint32) error
	RequestDataFromEpochCalled   func(identifier []byte) error
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
}

func (hrm *HeaderResolverStub) RequestDataFromEpoch(identifier []byte) error {
	if hrm.RequestDataFromEpochCalled != nil {
		return hrm.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

func (hrm *HeaderResolverStub) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrm.SetEpochHandlerCalled != nil {
		return hrm.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

func (hrm *HeaderResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if hrm.RequestDataFromHashCalled == nil {
		return nil
	}
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

func (hrm *HeaderResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	if hrm.ProcessReceivedMessageCalled == nil {
		return nil
	}
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *HeaderResolverStub) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	if hrm.RequestDataFromNonceCalled == nil {
		return nil
	}
	return hrm.RequestDataFromNonceCalled(nonce, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *HeaderResolverStub) IsInterfaceNil() bool {
	return hrm == nil
}
