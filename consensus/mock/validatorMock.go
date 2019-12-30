package mock

type ValidatorMock struct {
	pubKey  []byte
	address []byte
}

func NewValidatorMock(pubKey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{pubKey: pubKey, address: address}
}

func (vm *ValidatorMock) PubKey() []byte {
	return vm.pubKey
}

func (vm *ValidatorMock) Address() []byte {
	return vm.address
}

// IsInterfaceNil returns true if there is no value under the interface
func (vm *ValidatorMock) IsInterfaceNil() bool {
	if vm == nil {
		return true
	}
	return false
}
