package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/pkg/errors"
)

var errNotImplemented = errors.New("not implemented")

// HeaderResolverStub -
type HeaderResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64, epoch uint32) error
	RequestDataFromEpochCalled   func(identifier []byte) error
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
}

// RequestDataFromEpoch -
func (hrs *HeaderResolverStub) RequestDataFromEpoch(identifier []byte) error {
	if hrs.RequestDataFromEpochCalled != nil {
		return hrs.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

// SetEpochHandler -
func (hrs *HeaderResolverStub) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrs.SetEpochHandlerCalled != nil {
		return hrs.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

// RequestDataFromHash -
func (hrs *HeaderResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if hrs.RequestDataFromHashCalled != nil {
		return hrs.RequestDataFromHashCalled(hash, epoch)
	}

	return errNotImplemented
}

// ProcessReceivedMessage -
func (hrs *HeaderResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	if hrs.ProcessReceivedMessageCalled != nil {
		return hrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

// RequestDataFromNonce -
func (hrs *HeaderResolverStub) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	if hrs.RequestDataFromNonceCalled != nil {
		return hrs.RequestDataFromNonceCalled(nonce, epoch)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrs *HeaderResolverStub) IsInterfaceNil() bool {
	return hrs == nil
}
