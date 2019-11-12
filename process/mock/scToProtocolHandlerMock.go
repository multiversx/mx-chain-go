package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

type SCToProtocolStub struct {
	UpdateProtocolCalled func(body block.Body, nonce uint64) error
}

func (s *SCToProtocolStub) UpdateProtocol(body block.Body, nonce uint64) error {
	if s.UpdateProtocolCalled != nil {
		return s.UpdateProtocolCalled(body, nonce)
	}
	return nil
}

func (s *SCToProtocolStub) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
