package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type HeaderResolverMock struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	RequestDataFromNonceCalled   func(nonce uint64) error
}

func (hrm *HeaderResolverMock) RequestDataFromHash(hash []byte) error {
	if hrm.RequestDataFromHashCalled == nil {
		return nil
	}
	return hrm.RequestDataFromHashCalled(hash)
}

func (hrm *HeaderResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	if hrm.ProcessReceivedMessageCalled == nil {
		return nil
	}
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *HeaderResolverMock) RequestDataFromNonce(nonce uint64) error {
	if hrm.RequestDataFromNonceCalled == nil {
		return nil
	}
	return hrm.RequestDataFromNonceCalled(nonce)
}
