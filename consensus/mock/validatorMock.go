package mock

import (
	"math/big"
)

type ValidatorMock struct {
	stake  *big.Int
	rating int32
	pubKey []byte
}

func NewValidatorMock(stake *big.Int, rating int32, pubKey []byte) *ValidatorMock {
	return &ValidatorMock{stake: stake, rating: rating, pubKey: pubKey}
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

// IsInterfaceNil returns true if there is no value under the interface
func (vm *ValidatorMock) IsInterfaceNil() bool {
	if vm == nil {
		return true
	}
	return false
}
