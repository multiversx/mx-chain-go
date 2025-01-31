package txcachemocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
)

// TxGasHandler -
type TxGasHandler interface {
	MinGasPrice() uint64
	MaxGasLimitPerTx() uint64
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

// TxGasHandlerMock -
type TxGasHandlerMock struct {
	minGasLimit      uint64
	minGasPrice      uint64
	maxGasLimitPerTx uint64
	gasPerDataByte   uint64
	gasPriceModifier float64
}

// NewTxGasHandlerMock -
func NewTxGasHandlerMock() *TxGasHandlerMock {
	return &TxGasHandlerMock{
		minGasLimit:      50000,
		minGasPrice:      1000000000,
		maxGasLimitPerTx: 600000000,
		gasPerDataByte:   1500,
		gasPriceModifier: 0.01,
	}
}

// WithMinGasLimit -
func (ghm *TxGasHandlerMock) WithMinGasLimit(minGasLimit uint64) *TxGasHandlerMock {
	ghm.minGasLimit = minGasLimit
	return ghm
}

// WithMinGasPrice -
func (ghm *TxGasHandlerMock) WithMinGasPrice(minGasPrice uint64) *TxGasHandlerMock {
	ghm.minGasPrice = minGasPrice
	return ghm
}

// MinGasPrice -
func (ghm *TxGasHandlerMock) MinGasPrice() uint64 {
	return ghm.minGasPrice
}

// WithGasPriceModifier -
func (ghm *TxGasHandlerMock) WithGasPriceModifier(gasPriceModifier float64) *TxGasHandlerMock {
	ghm.gasPriceModifier = gasPriceModifier
	return ghm
}

// MaxGasLimitPerTx -
func (ghm *TxGasHandlerMock) MaxGasLimitPerTx() uint64 {
	return ghm.maxGasLimitPerTx
}

// ComputeTxFee -
func (ghm *TxGasHandlerMock) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	dataLength := uint64(len(tx.GetData()))
	gasPriceForMovement := tx.GetGasPrice()
	gasPriceForProcessing := uint64(float64(gasPriceForMovement) * ghm.gasPriceModifier)

	gasLimitForMovement := ghm.minGasLimit + dataLength*ghm.gasPerDataByte
	if tx.GetGasLimit() < gasLimitForMovement {
		return big.NewInt(0)
	}

	gasLimitForProcessing := tx.GetGasLimit() - gasLimitForMovement
	feeForMovement := core.SafeMul(gasPriceForMovement, gasLimitForMovement)
	feeForProcessing := core.SafeMul(gasPriceForProcessing, gasLimitForProcessing)
	fee := big.NewInt(0).Add(feeForMovement, feeForProcessing)
	return fee
}

// IsInterfaceNil -
func (ghm *TxGasHandlerMock) IsInterfaceNil() bool {
	return ghm == nil
}
