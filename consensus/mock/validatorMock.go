package mock

import (
	"math/big"
)

type ValidatorMock struct {
	stake  big.Int
	rating int32
	pubKey []byte
}

func NewValidatorMock(stake big.Int, rating int32, pubKey []byte) *ValidatorMock {
	return &ValidatorMock{stake: stake, rating: rating, pubKey: pubKey}
}

func (vm *ValidatorMock) Stake() big.Int {
	panic("implement me")
}

func (vm *ValidatorMock) Rating() int32 {
	panic("implement me")
}

func (vm *ValidatorMock) PubKey() []byte {
	panic("implement me")
}
