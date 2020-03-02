package mock

import "math/big"

// ValidatorSettingsStub -
type ValidatorSettingsStub struct {
}

// UnBondPeriod -
func (v *ValidatorSettingsStub) UnBondPeriod() uint64 {
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
