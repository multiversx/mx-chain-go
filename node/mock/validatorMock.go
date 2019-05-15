package mock

import (
	"math/big"
)

type ValidatorMock struct {
	StakeField  *big.Int
	RatingField int32
	PubKeyField []byte
}

func (vm *ValidatorMock) Stake() *big.Int {
	return vm.StakeField
}

func (vm *ValidatorMock) Rating() int32 {
	return vm.RatingField
}

func (vm *ValidatorMock) PubKey() []byte {
	return vm.PubKeyField
}
