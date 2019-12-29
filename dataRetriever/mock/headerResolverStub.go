package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/pkg/errors"
)

var errNotImplemented = errors.New("not implemented")

type HeaderResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64) error
	RequestDataFromEpochCalled   func(identifier []byte) error
	SetEpochHandlerCalled        func(epochHandler dataRetriever.EpochHandler) error
}

func (hrs *HeaderResolverStub) RequestDataFromEpoch(identifier []byte) error {
	if hrs.RequestDataFromEpochCalled != nil {
		return hrs.RequestDataFromEpochCalled(identifier)
	}
	return nil
}

func (hrs *HeaderResolverStub) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if hrs.SetEpochHandlerCalled != nil {
		return hrs.SetEpochHandlerCalled(epochHandler)
	}
	return nil
}

func (hrs *HeaderResolverStub) RequestDataFromHash(hash []byte) error {
	if hrs.RequestDataFromHashCalled != nil {
		return hrs.RequestDataFromHashCalled(hash)
	}

	return errNotImplemented
}

func (hrs *HeaderResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	if hrs.ProcessReceivedMessageCalled != nil {
		return hrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

func (hrs *HeaderResolverStub) RequestDataFromNonce(nonce uint64) error {
	if hrs.RequestDataFromNonceCalled != nil {
		return hrs.RequestDataFromNonceCalled(nonce)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrs *HeaderResolverStub) IsInterfaceNil() bool {
	if hrs == nil {
		return true
	}
	return false
}
