package mock

import (
	"math/big"
)

type ValidatorMock struct {
	stake   *big.Int
	rating  int32
	pubKey  []byte
	address []byte
}

func NewValidatorMock(stake *big.Int, rating int32, pubKey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{stake: stake, rating: rating, pubKey: pubKey, address: address}
}

func (vm *ValidatorMock) Stake() *big.Int {
	return vm.stake
}

func (vm *ValidatorMock) Rating() int32 {
	return vm.rating
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
