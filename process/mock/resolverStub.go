package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type ResolverStub struct {
	RequestHashCalled func(hash []byte) error
	ValidateCalled    func(message p2p.MessageP2P) error
}

func (rs *ResolverStub) RequestHash(hash []byte) error {
	return rs.RequestHashCalled(hash)
}

func (rs *ResolverStub) Validate(message p2p.MessageP2P) error {
	return rs.ValidateCalled(message)
}
