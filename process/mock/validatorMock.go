package mock

import "sync"

// ValidatorMock -
type ValidatorMock struct {
	pubKey     []byte
	address    []byte
	chances    uint32
	mutChances sync.RWMutex

	PubKeyCalled  func() []byte
	AddressCalled func() []byte
}

// NewValidatorMock -
func NewValidatorMock(pubkey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{
		pubKey:  pubkey,
		address: address,
	}
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

// Chances -
func (vm *ValidatorMock) Chances() uint32 {
	vm.mutChances.RLock()
	defer vm.mutChances.RUnlock()

	return vm.chances
}

// SetChances -
func (vm *ValidatorMock) SetChances(chances uint32) {
	vm.mutChances.Lock()
	vm.chances = chances
	vm.mutChances.Unlock()
}
