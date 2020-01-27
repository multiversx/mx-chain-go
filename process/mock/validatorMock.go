package mock

import (
	"math/big"
)

type ValidatorMock struct {
	stake   *big.Int
	rating  int32
	pubKey  []byte
	address []byte

	PubKeyCalled  func() []byte
	AddressCalled func() []byte
}

func (vm *ValidatorMock) Stake() *big.Int {
	return vm.stake
}

func (vm *ValidatorMock) Rating() int32 {
	return vm.rating
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
