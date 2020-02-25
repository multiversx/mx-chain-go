package mock

// ValidatorMock -
type ValidatorMock struct {
	pubKey  []byte
	address []byte

	PubKeyCalled  func() []byte
	AddressCalled func() []byte
}

// PubKey -
func (vm *ValidatorMock) PubKey() []byte {
	if vm.PubKeyCalled != nil {
		return vm.PubKeyCalled()
	}
	return vm.pubKey
}

// Address -
func (vm *ValidatorMock) Address() []byte {
	if vm.AddressCalled != nil {
		return vm.AddressCalled()
	}
	return vm.address
}
