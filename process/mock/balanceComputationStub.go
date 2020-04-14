package mock

import (
	"math/big"
)

// BalanceComputationStub -
type BalanceComputationStub struct {
	InitCalled                  func()
	SetBalanceToAddressCalled   func(address []byte, value *big.Int)
	GetBalanceOfAddressCalled   func(address []byte) (*big.Int, error)
	AddBalanceToAddressCalled   func(address []byte, value *big.Int) bool
	SubBalanceFromAddressCalled func(address []byte, value *big.Int) (bool, bool)
	HasAddressBalanceSetCalled  func(address []byte) bool
	IsBalanceInAddressCalled    func(address []byte, value *big.Int) bool
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

// GetBalanceOfAddress -
func (bcs *BalanceComputationStub) GetBalanceOfAddress(address []byte) (*big.Int, error) {
	if bcs.GetBalanceOfAddressCalled != nil {
		return bcs.GetBalanceOfAddressCalled(address)
	}

	return big.NewInt(0), nil
}

// AddBalanceToAddress -
func (bcs *BalanceComputationStub) AddBalanceToAddress(address []byte, value *big.Int) bool {
	if bcs.AddBalanceToAddressCalled != nil {
		return bcs.AddBalanceToAddressCalled(address, value)
	}

	return true
}

// SubBalanceFromAddress -
func (bcs *BalanceComputationStub) SubBalanceFromAddress(address []byte, value *big.Int) (bool, bool) {
	if bcs.SubBalanceFromAddressCalled != nil {
		return bcs.SubBalanceFromAddressCalled(address, value)
	}

	return true, true
}

// HasAddressBalanceSet -
func (bcs *BalanceComputationStub) HasAddressBalanceSet(address []byte) bool {
	if bcs.HasAddressBalanceSetCalled != nil {
		return bcs.HasAddressBalanceSetCalled(address)
	}

	return false
}

// IsBalanceInAddress -
func (bcs *BalanceComputationStub) IsBalanceInAddress(address []byte, value *big.Int) bool {
	if bcs.IsBalanceInAddressCalled != nil {
		return bcs.IsBalanceInAddressCalled(address, value)
	}

	return true
}

// IsInterfaceNil -
func (bcs *BalanceComputationStub) IsInterfaceNil() bool {
	return bcs == nil
}
