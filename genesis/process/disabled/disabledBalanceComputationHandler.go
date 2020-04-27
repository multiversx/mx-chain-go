package disabled

import "math/big"

// BalanceComputationHandler -
type BalanceComputationHandler struct {
}

// Init -
func (b *BalanceComputationHandler) Init() {
}

// SetBalanceToAddress -
func (b *BalanceComputationHandler) SetBalanceToAddress(_ []byte, _ *big.Int) {
}

// AddBalanceToAddress -
func (b *BalanceComputationHandler) AddBalanceToAddress(_ []byte, _ *big.Int) bool {
	return true
}

// SubBalanceFromAddress -
func (b *BalanceComputationHandler) SubBalanceFromAddress(_ []byte, _ *big.Int) bool {
	return true
}

// IsAddressSet -
func (b *BalanceComputationHandler) IsAddressSet(_ []byte) bool {
	return true
}

// AddressHasEnoughBalance -
func (b *BalanceComputationHandler) AddressHasEnoughBalance(_ []byte, _ *big.Int) bool {
	return true
}

// IsInterfaceNil -
func (b *BalanceComputationHandler) IsInterfaceNil() bool {
	return b == nil
}
