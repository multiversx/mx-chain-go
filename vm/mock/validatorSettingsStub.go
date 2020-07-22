package mock

import "math/big"

// ValidatorSettingsStub -
type ValidatorSettingsStub struct {
	MinStepValueCalled             func() *big.Int
	TotalSupplyCalled              func() *big.Int
	NumNodesCalled                 func() uint32
	AuctionEnableNonceCalled       func() uint64
	StakeEnableNonceCalled         func() uint64
	UnBondPeriodCalled             func() uint64
	StakeValueCalled               func() *big.Int
	UnJailValueCalled              func() *big.Int
	NumRoundsWithoutBleedCalled    func() uint64
	BleedPercentagePerRoundCalled  func() float64
	MaximumPercentageToBleedCalled func() float64
}

// NumRoundsWithoutBleed -
func (v *ValidatorSettingsStub) NumRoundsWithoutBleed() uint64 {
	if v.NumRoundsWithoutBleedCalled != nil {
		return v.NumRoundsWithoutBleedCalled()
	}
	return 0
}

// BleedPercentagePerRound -
func (v *ValidatorSettingsStub) BleedPercentagePerRound() float64 {
	if v.BleedPercentagePerRoundCalled != nil {
		return v.BleedPercentagePerRoundCalled()
	}
	return 0
}

// MaximumPercentageToBleed -
func (v *ValidatorSettingsStub) MaximumPercentageToBleed() float64 {
	if v.MaximumPercentageToBleedCalled != nil {
		return v.MaximumPercentageToBleedCalled()
	}
	return 0
}

// MinStepValue -
func (v *ValidatorSettingsStub) MinStepValue() *big.Int {
	if v.MinStepValueCalled != nil {
		return v.MinStepValueCalled()
	}
	return big.NewInt(100000)
}

// UnJailValue -
func (v *ValidatorSettingsStub) UnJailValue() *big.Int {
	if v.UnJailValueCalled != nil {
		return v.UnJailValueCalled()
	}
	return big.NewInt(100000)
}

// GenesisTotalSupply -
func (v *ValidatorSettingsStub) GenesisTotalSupply() *big.Int {
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

// AuctionEnableNonce -
func (v *ValidatorSettingsStub) AuctionEnableNonce() uint64 {
	if v.AuctionEnableNonceCalled != nil {
		return v.AuctionEnableNonceCalled()
	}
	return 1000000
}

// StakeEnableNonce -
func (v *ValidatorSettingsStub) StakeEnableNonce() uint64 {
	if v.StakeEnableNonceCalled != nil {
		return v.StakeEnableNonceCalled()
	}
	return 10000000
}

// UnBondPeriod -
func (v *ValidatorSettingsStub) UnBondPeriod() uint64 {
	if v.UnBondPeriodCalled != nil {
		return v.UnBondPeriodCalled()
	}
	return 100000
}

// GenesisNodePrice -
func (v *ValidatorSettingsStub) GenesisNodePrice() *big.Int {
	if v.StakeValueCalled != nil {
		return v.StakeValueCalled()
	}
	return big.NewInt(10000000)
}

// IsInterfaceNil -
func (v *ValidatorSettingsStub) IsInterfaceNil() bool {
	return v == nil
}
