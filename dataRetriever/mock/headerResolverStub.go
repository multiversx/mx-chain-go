package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/pkg/errors"
)

var errNotImplemented = errors.New("not implemented")

type HeaderResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64) error
}

func (hrs *HeaderResolverStub) RequestDataFromHash(hash []byte) error {
	if hrs.RequestDataFromHashCalled != nil {
		return hrs.RequestDataFromHashCalled(hash)
	}

	return errNotImplemented
}

func (hrs *HeaderResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
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
