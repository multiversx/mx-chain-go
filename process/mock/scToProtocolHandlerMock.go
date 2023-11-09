package mock

import "github.com/multiversx/mx-chain-core-go/data/block"

// SCToProtocolStub -
type SCToProtocolStub struct {
	UpdateProtocolCalled func(body *block.Body, nonce uint64) error
}

// UpdateProtocol -
func (s *SCToProtocolStub) UpdateProtocol(body *block.Body, nonce uint64) error {
	if s.UpdateProtocolCalled != nil {
		return s.UpdateProtocolCalled(body, nonce)
	}
	return nil
}

// IsInterfaceNil -
func (s *SCToProtocolStub) IsInterfaceNil() bool {
	return s == nil
}
