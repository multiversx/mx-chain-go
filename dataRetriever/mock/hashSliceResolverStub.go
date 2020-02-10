package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// HashSliceResolverStub -
type HashSliceResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
}

// RequestDataFromHash -
func (hsrs *HashSliceResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if hsrs.RequestDataFromHashCalled != nil {
		return hsrs.RequestDataFromHashCalled(hash, epoch)
	}

	return errNotImplemented
}

// ProcessReceivedMessage -
func (hsrs *HashSliceResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	if hsrs.ProcessReceivedMessageCalled != nil {
		return hsrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

// RequestDataFromHashArray -
func (hsrs *HashSliceResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if hsrs.RequestDataFromHashArrayCalled != nil {
		return hsrs.RequestDataFromHashArrayCalled(hashes, epoch)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (hsrs *HashSliceResolverStub) IsInterfaceNil() bool {
	return hsrs == nil
}
