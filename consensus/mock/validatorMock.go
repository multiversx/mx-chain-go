package mock

// ValidatorMock -
type ValidatorMock struct {
	pubKey  []byte
	address []byte
}

// NewValidatorMock -
func NewValidatorMock(pubKey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{pubKey: pubKey, address: address}
}

// PubKey -
func (vm *ValidatorMock) PubKey() []byte {
	return vm.pubKey
}

// Address -
func (vm *ValidatorMock) Address() []byte {
	return vm.address
}

// IsInterfaceNil returns true if there is no value under the interface
func (vm *ValidatorMock) IsInterfaceNil() bool {
	return vm == nil
}
