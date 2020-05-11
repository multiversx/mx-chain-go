package disabled

import "math/big"

// BalanceComputationHandler implements BalanceComputationHandler interface but does nothing as it is a disabled component
type BalanceComputationHandler struct {
}

// Init does nothing as it is a disabled component
func (b *BalanceComputationHandler) Init() {
}

// SetBalanceToAddress does nothing as it is a disabled component
func (b *BalanceComputationHandler) SetBalanceToAddress(_ []byte, _ *big.Int) {
}

// AddBalanceToAddress returns true as it is a disabled component
func (b *BalanceComputationHandler) AddBalanceToAddress(_ []byte, _ *big.Int) bool {
	return true
}

// SubBalanceFromAddress returns true as it is a disabled component
func (b *BalanceComputationHandler) SubBalanceFromAddress(_ []byte, _ *big.Int) bool {
	return true
}

// IsAddressSet returns true as it is a disabled component
func (b *BalanceComputationHandler) IsAddressSet(_ []byte) bool {
	return true
}

// AddressHasEnoughBalance returns true as it is a disabled component
func (b *BalanceComputationHandler) AddressHasEnoughBalance(_ []byte, _ *big.Int) bool {
	return true
}

// IsInterfaceNil returns true if underlying object is nil
func (b *BalanceComputationHandler) IsInterfaceNil() bool {
	return b == nil
}
