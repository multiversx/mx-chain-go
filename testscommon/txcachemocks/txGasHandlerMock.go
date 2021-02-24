package txcachemocks

import "github.com/ElrondNetwork/elrond-go/process"

// TxGasHandler -
type TxGasHandler interface {
	SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64
	GasPriceForMove(tx process.TransactionWithFeeHandler) uint64
	MinGasPrice() uint64
	MinGasLimit() uint64
	MinGasPriceForProcessing() uint64
	IsInterfaceNil() bool
}

// TxGasHandlerMock -
type TxGasHandlerMock struct {
	MinimumGasMove       uint64
	MinimumGasPrice      uint64
	GasProcessingDivisor uint64
}

// SplitTxGasInCategories -
func (ghm *TxGasHandlerMock) SplitTxGasInCategories(tx process.TransactionWithFeeHandler) (uint64, uint64) {
	moveGas := ghm.MinimumGasMove
	return moveGas, tx.GetGasLimit() - moveGas
}

// GasPriceForProcessing -
func (ghm *TxGasHandlerMock) GasPriceForProcessing(tx process.TransactionWithFeeHandler) uint64 {
	return tx.GetGasPrice() / ghm.GasProcessingDivisor
}

// GasPriceForMove -
func (ghm *TxGasHandlerMock) GasPriceForMove(tx process.TransactionWithFeeHandler) uint64 {
	return tx.GetGasPrice()
}

// MinGasPrice -
func (ghm *TxGasHandlerMock) MinGasPrice() uint64 {
	return ghm.MinimumGasPrice
}

// MinGasLimit -
func (ghm *TxGasHandlerMock) MinGasLimit() uint64 {
	return ghm.MinimumGasMove
}

// MinGasPriceProcessing -
func (ghm *TxGasHandlerMock) MinGasPriceForProcessing() uint64 {
	return ghm.MinimumGasPrice / ghm.GasProcessingDivisor
}

// IsInterfaceNil -
func (ghm *TxGasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
