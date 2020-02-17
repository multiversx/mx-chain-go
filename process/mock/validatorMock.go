package mock

import (
	"math/big"
)

// ValidatorMock -
type ValidatorMock struct {
	stake   *big.Int
	rating  int32
	pubKey  []byte
	address []byte

	PubKeyCalled  func() []byte
	AddressCalled func() []byte
}

// Stake -
func (vm *ValidatorMock) Stake() *big.Int {
	return vm.stake
}

// Rating -
func (vm *ValidatorMock) Rating() int32 {
	return vm.rating
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
