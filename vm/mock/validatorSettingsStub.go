package mock

import "math/big"

// ValidatorSettingsStub -
type ValidatorSettingsStub struct {
}

// UnBoundPeriod -
func (v *ValidatorSettingsStub) UnBoundPeriod() uint64 {
	return 10
}

// GenesisNodePrice -
func (v *ValidatorSettingsStub) GenesisNodePrice() *big.Int {
	return big.NewInt(10)
}

// IsInterfaceNil -
func (v *ValidatorSettingsStub) IsInterfaceNil() bool {
	return v == nil
}
