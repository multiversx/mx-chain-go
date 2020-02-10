package mock

import "math/big"

// ValidatorSettingsStub -
type ValidatorSettingsStub struct {
	MinStepValueCalled   func() *big.Int
	TotalSupplyCalled    func() *big.Int
	NumNodesCalled       func() uint32
	AuctionEnabledCalled func() bool
	UnBondPeriodCalled   func() uint64
	StakeValueCalled     func() *big.Int
}

// MinStepValue -
func (v *ValidatorSettingsStub) MinStepValue() *big.Int {
	if v.MinStepValueCalled != nil {
		return v.MinStepValueCalled()
	}
	return big.NewInt(100000)
}

// TotalSupply -
func (v *ValidatorSettingsStub) TotalSupply() *big.Int {
	if v.TotalSupplyCalled != nil {
		return v.TotalSupplyCalled()
	}
	return big.NewInt(100000000000)
}

// NumNodes -
func (v *ValidatorSettingsStub) NumNodes() uint32 {
	if v.NumNodesCalled != nil {
		return v.NumNodesCalled()
	}
	return 10
}

// AuctionEnabled -
func (v *ValidatorSettingsStub) AuctionEnabled() bool {
	if v.AuctionEnabledCalled != nil {
		return v.AuctionEnabledCalled()
	}
	return false
}

// UnBondPeriod -
func (v *ValidatorSettingsStub) UnBondPeriod() uint64 {
	if v.UnBondPeriodCalled != nil {
		return v.UnBondPeriodCalled()
	}
	return 100000
}

// StakeValue -
func (v *ValidatorSettingsStub) StakeValue() *big.Int {
	if v.StakeValueCalled != nil {
		return v.StakeValueCalled()
	}
	return big.NewInt(10000000)
}

// IsInterfaceNil -
func (v *ValidatorSettingsStub) IsInterfaceNil() bool {
	return v == nil
}
