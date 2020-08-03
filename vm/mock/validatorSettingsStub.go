package mock

import "math/big"

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	TotalSupplyCalled func() *big.Int
}

// GenesisTotalSupply -
func (v *EconomicsHandlerStub) GenesisTotalSupply() *big.Int {
	if v.TotalSupplyCalled != nil {
		return v.TotalSupplyCalled()
	}
	return big.NewInt(100000000000)
}

// IsInterfaceNil -
func (v *EconomicsHandlerStub) IsInterfaceNil() bool {
	return v == nil
}
