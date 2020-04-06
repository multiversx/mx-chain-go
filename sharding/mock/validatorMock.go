package mock

import (
	"sync"
)

// ValidatorMock -
type ValidatorMock struct {
	pubKey     []byte
	address    []byte
	chances    uint32
	mutChances sync.RWMutex
}

// NewValidatorMock -
func NewValidatorMock(pubKey []byte, address []byte, chances uint32) *ValidatorMock {
	return &ValidatorMock{pubKey: pubKey, address: address, chances: chances}
}

// PubKey -
func (vm *ValidatorMock) PubKey() []byte {
	return vm.pubKey
}

// Address -
func (vm *ValidatorMock) Address() []byte {
	return vm.address
}

// Chances -
func (vm *ValidatorMock) Chances() uint32 {
	vm.mutChances.RLock()
	defer vm.mutChances.RUnlock()

	return vm.chances
}

// Clone clones the validator
func (vm *ValidatorMock) Clone() (interface{}, error) {
	v2 := *vm
	return &v2, nil
}

// SetChances -
func (vm *ValidatorMock) SetChances(chances uint32) {
	vm.mutChances.Lock()
	vm.chances = chances
	vm.mutChances.Unlock()
}
