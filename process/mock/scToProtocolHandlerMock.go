package mock

import "github.com/ElrondNetwork/elrond-go/data/block"

type SCToProtocolMock struct {
	UpdateProtocolCalled func(body block.Body, nonce uint64) error
}

func (s *SCToProtocolMock) UpdateProtocol(body block.Body, nonce uint64) error {
	if s.UpdateProtocolCalled != nil {
		return s.UpdateProtocolCalled(body, nonce)
	}
	return nil
}

func (s *SCToProtocolMock) IsInterfaceNil() bool {
	if s == nil {
		return true
	}
	return false
}
