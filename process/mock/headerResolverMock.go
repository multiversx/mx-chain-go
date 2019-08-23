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
	return hrm.RequestDataFromHashCalled(hash)
}

func (hrm *HeaderResolverMock) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *HeaderResolverMock) RequestDataFromNonce(nonce uint64) error {
	return hrm.RequestDataFromNonceCalled(nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *HeaderResolverMock) IsInterfaceNil() bool {
	if hrm == nil {
		return true
	}
	return false
}
