package mock

import (
	"math/big"
)

// BalanceComputationStub -
type BalanceComputationStub struct {
	InitCalled                    func()
	SetBalanceToAddressCalled     func(address []byte, value *big.Int)
	AddBalanceToAddressCalled     func(address []byte, value *big.Int) bool
	SubBalanceFromAddressCalled   func(address []byte, value *big.Int) bool
	IsAddressSetCalled            func(address []byte) bool
	AddressHasEnoughBalanceCalled func(address []byte, value *big.Int) bool
}

// Init -
func (bcs *BalanceComputationStub) Init() {
	if bcs.InitCalled != nil {
		bcs.InitCalled()
	}
}

// SetBalanceToAddress -
func (bcs *BalanceComputationStub) SetBalanceToAddress(address []byte, value *big.Int) {
	if bcs.SetBalanceToAddressCalled != nil {
		bcs.SetBalanceToAddressCalled(address, value)
	}
}

// AddBalanceToAddress -
func (bcs *BalanceComputationStub) AddBalanceToAddress(address []byte, value *big.Int) bool {
	if bcs.AddBalanceToAddressCalled != nil {
		return bcs.AddBalanceToAddressCalled(address, value)
	}

	return true
}

// SubBalanceFromAddress -
func (bcs *BalanceComputationStub) SubBalanceFromAddress(address []byte, value *big.Int) bool {
	if bcs.SubBalanceFromAddressCalled != nil {
		return bcs.SubBalanceFromAddressCalled(address, value)
	}

	return true
}

// IsAddressSet -
func (bcs *BalanceComputationStub) IsAddressSet(address []byte) bool {
	if bcs.IsAddressSetCalled != nil {
		return bcs.IsAddressSetCalled(address)
	}

	return false
}

// AddressHasEnoughBalance -
func (bcs *BalanceComputationStub) AddressHasEnoughBalance(address []byte, value *big.Int) bool {
	if bcs.AddressHasEnoughBalanceCalled != nil {
		return bcs.AddressHasEnoughBalanceCalled(address, value)
	}

	return true
}

// IsInterfaceNil -
func (bcs *BalanceComputationStub) IsInterfaceNil() bool {
	return bcs == nil
}
