package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type HashSliceResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	RequestDataFromHashArrayCalled func(hashes [][]byte) error
}

func (hsrs *HashSliceResolverStub) RequestDataFromHash(hash []byte) error {
	if hsrs.RequestDataFromHashCalled != nil {
		return hsrs.RequestDataFromHashCalled(hash)
	}

	return errNotImplemented
}

func (hsrs *HashSliceResolverStub) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if hsrs.ProcessReceivedMessageCalled != nil {
		return hsrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

func (hsrs *HashSliceResolverStub) RequestDataFromHashArray(hashes [][]byte) error {
	if hsrs.RequestDataFromHashArrayCalled != nil {
		return hsrs.RequestDataFromHashArrayCalled(hashes)
	}

	return errNotImplemented
}
