package mock

type ValidatorMock struct {
	pubKey  []byte
	address []byte

	PubKeyCalled  func() []byte
	AddressCalled func() []byte
}

func NewValidatorMock(pubKey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{pubKey: pubKey, address: address}
}

func (vm *ValidatorMock) PubKey() []byte {
	if vm.PubKeyCalled != nil {
		return vm.PubKeyCalled()
	}
	return vm.pubKey
}

func (vm *ValidatorMock) Address() []byte {
	if vm.AddressCalled != nil {
		return vm.AddressCalled()
	}
	return vm.address
}
