package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

type HashSliceResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
}

func (hsrs *HashSliceResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if hsrs.RequestDataFromHashCalled != nil {
		return hsrs.RequestDataFromHashCalled(hash, epoch)
	}

	return errNotImplemented
}

func (hsrs *HashSliceResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	if hsrs.ProcessReceivedMessageCalled != nil {
		return hsrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

func (hsrs *HashSliceResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if hsrs.RequestDataFromHashArrayCalled != nil {
		return hsrs.RequestDataFromHashArrayCalled(hashes, epoch)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (hsrs *HashSliceResolverStub) IsInterfaceNil() bool {
	if hsrs == nil {
		return true
	}
	return false
}
