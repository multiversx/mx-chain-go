package mock

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
