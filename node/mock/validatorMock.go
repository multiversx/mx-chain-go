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
}

// NewValidatorMock -
func NewValidatorMock(stake *big.Int, rating int32, pubKey []byte, address []byte) *ValidatorMock {
	return &ValidatorMock{stake: stake, rating: rating, pubKey: pubKey, address: address}
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
	return vm.pubKey
}

// Address -
func (vm *ValidatorMock) Address() []byte {
	return vm.address
}
