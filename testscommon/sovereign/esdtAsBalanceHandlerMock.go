package sovereign

import (
	"math/big"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ESDTAsBalanceHandlerMock -
type ESDTAsBalanceHandlerMock struct {
	GetBalanceCalled     func(accountDataHandler vmcommon.AccountDataHandler) *big.Int
	AddToBalanceCalled   func(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error
	SubFromBalanceCalled func(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error
}

// GetBalance -
func (mock *ESDTAsBalanceHandlerMock) GetBalance(accountDataHandler vmcommon.AccountDataHandler) *big.Int {
	if mock.GetBalanceCalled != nil {
		return mock.GetBalanceCalled(accountDataHandler)
	}
	return nil
}

// AddToBalance -
func (mock *ESDTAsBalanceHandlerMock) AddToBalance(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
	if mock.AddToBalanceCalled != nil {
		return mock.AddToBalanceCalled(accountDataHandler, value)
	}
	return nil
}

// SubFromBalance -
func (mock *ESDTAsBalanceHandlerMock) SubFromBalance(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
	if mock.SubFromBalanceCalled != nil {
		return mock.SubFromBalanceCalled(accountDataHandler, value)
	}
	return nil
}

// IsInterfaceNil -
func (mock *ESDTAsBalanceHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
