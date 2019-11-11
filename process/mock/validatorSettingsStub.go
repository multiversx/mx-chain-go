package mock

import "math/big"

type ValidatorSettingsStub struct {
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
