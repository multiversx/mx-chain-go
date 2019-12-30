package mock

import "math/big"

type ValidatorSettingsStub struct {
}

func (v *ValidatorSettingsStub) MinStepValue() *big.Int {
	return big.NewInt(1000)
}

func (v *ValidatorSettingsStub) TotalSupply() *big.Int {
	return big.NewInt(10000000)
}

func (v *ValidatorSettingsStub) NumNodes() uint32 {
	return 100
}

func (v *ValidatorSettingsStub) AuctionEnabled() bool {
	return false
}

func (v *ValidatorSettingsStub) UnBoundPeriod() uint64 {
	return 10
}

func (v *ValidatorSettingsStub) StakeValue() *big.Int {
	return big.NewInt(10)
}

func (v *ValidatorSettingsStub) IsInterfaceNil() bool {
	return v == nil
}
