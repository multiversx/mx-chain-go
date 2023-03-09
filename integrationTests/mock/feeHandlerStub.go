package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// FeeHandlerStub -
type FeeHandlerStub struct {
	SetMaxGasLimitPerBlockCalled                   func(maxGasLimitPerBlock uint64)
	SetMinGasPriceCalled                           func(minGasPrice uint64)
	SetMinGasLimitCalled                           func(minGasLimit uint64)
	MaxGasLimitPerBlockCalled                      func() uint64
	MaxGasLimitPerMiniBlockCalled                  func() uint64
	MaxGasLimitPerBlockForSafeCrossShardCalled     func() uint64
	MaxGasLimitPerMiniBlockForSafeCrossShardCalled func() uint64
	MaxGasLimitPerTxCalled                         func() uint64
	ComputeGasLimitCalled                          func(tx data.TransactionWithFeeHandler) uint64
	ComputeMoveBalanceFeeCalled                    func(tx data.TransactionWithFeeHandler) *big.Int
	ComputeTxFeeCalled                             func(tx data.TransactionWithFeeHandler) *big.Int
	CheckValidityTxValuesCalled                    func(tx data.TransactionWithFeeHandler) error
	DeveloperPercentageCalled                      func() float64
	MinGasPriceCalled                              func() uint64
	GasPriceModifierCalled                         func() float64
	ComputeFeeForProcessingCalled                  func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int
	GenesisTotalSupplyCalled                       func() *big.Int
	SplitTxGasInCategoriesCalled                   func(tx data.TransactionWithFeeHandler) (uint64, uint64)
	GasPriceForProcessingCalled                    func(tx data.TransactionWithFeeHandler) uint64
	GasPriceForMoveCalled                          func(tx data.TransactionWithFeeHandler) uint64
	MinGasPriceForProcessingCalled                 func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled   func(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled               func(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeGasLimitBasedOnBalanceCalled            func(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error)
}

// ComputeGasLimitBasedOnBalance -
func (fhs *FeeHandlerStub) ComputeGasLimitBasedOnBalance(tx data.TransactionWithFeeHandler, balance *big.Int) (uint64, error) {
	if fhs.ComputeGasLimitBasedOnBalanceCalled != nil {
		return fhs.ComputeGasLimitBasedOnBalanceCalled(tx, balance)
	}
	return 0, nil
}

// ComputeFeeForProcessing -
func (fhs *FeeHandlerStub) ComputeFeeForProcessing(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	if fhs.ComputeFeeForProcessingCalled != nil {
		return fhs.ComputeFeeForProcessingCalled(tx, gasToUse)
	}
	return big.NewInt(0)
}

// GasPriceModifier -
func (fhs *FeeHandlerStub) GasPriceModifier() float64 {
	if fhs.GasPriceModifierCalled != nil {
		return fhs.GasPriceModifierCalled()
	}
	return 1.0
}

// MinGasPrice -
func (fhs *FeeHandlerStub) MinGasPrice() uint64 {
	if fhs.MinGasPriceCalled != nil {
		return fhs.MinGasPriceCalled()
	}
	return 0
}

// MinGasLimit will return min gas limit
func (fhs *FeeHandlerStub) MinGasLimit() uint64 {
	return 0
}

// DeveloperPercentage -
func (fhs *FeeHandlerStub) DeveloperPercentage() float64 {
	if fhs.DeveloperPercentageCalled != nil {
		return fhs.DeveloperPercentageCalled()
	}

	return 0.0
}

// GasPerDataByte -
func (fhs *FeeHandlerStub) GasPerDataByte() uint64 {
	return 0
}

// SetMaxGasLimitPerBlock -
func (fhs *FeeHandlerStub) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64) {
	fhs.SetMaxGasLimitPerBlockCalled(maxGasLimitPerBlock)
}

// SetMinGasPrice -
func (fhs *FeeHandlerStub) SetMinGasPrice(minGasPrice uint64) {
	fhs.SetMinGasPriceCalled(minGasPrice)
}

// SetMinGasLimit -
func (fhs *FeeHandlerStub) SetMinGasLimit(minGasLimit uint64) {
	fhs.SetMinGasLimitCalled(minGasLimit)
}

// MaxGasLimitPerBlock -
func (fhs *FeeHandlerStub) MaxGasLimitPerBlock(uint32) uint64 {
	if fhs.MaxGasLimitPerBlockCalled != nil {
		return fhs.MaxGasLimitPerBlockCalled()
	}
	return 0
}

// MaxGasLimitPerMiniBlock -
func (fhs *FeeHandlerStub) MaxGasLimitPerMiniBlock(uint32) uint64 {
	if fhs.MaxGasLimitPerMiniBlockCalled != nil {
		return fhs.MaxGasLimitPerMiniBlockCalled()
	}
	return 0
}

// MaxGasLimitPerBlockForSafeCrossShard -
func (fhs *FeeHandlerStub) MaxGasLimitPerBlockForSafeCrossShard() uint64 {
	if fhs.MaxGasLimitPerBlockForSafeCrossShardCalled != nil {
		return fhs.MaxGasLimitPerBlockForSafeCrossShardCalled()
	}
	return 0
}

// MaxGasLimitPerMiniBlockForSafeCrossShard -
func (fhs *FeeHandlerStub) MaxGasLimitPerMiniBlockForSafeCrossShard() uint64 {
	if fhs.MaxGasLimitPerMiniBlockForSafeCrossShardCalled != nil {
		return fhs.MaxGasLimitPerMiniBlockForSafeCrossShardCalled()
	}
	return 0
}

// MaxGasLimitPerTx -
func (fhs *FeeHandlerStub) MaxGasLimitPerTx() uint64 {
	if fhs.MaxGasLimitPerTxCalled != nil {
		return fhs.MaxGasLimitPerTxCalled()
	}
	return 0
}

// ComputeGasLimit -
func (fhs *FeeHandlerStub) ComputeGasLimit(tx data.TransactionWithFeeHandler) uint64 {
	if fhs.ComputeGasLimitCalled != nil {
		return fhs.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeMoveBalanceFee -
func (fhs *FeeHandlerStub) ComputeMoveBalanceFee(tx data.TransactionWithFeeHandler) *big.Int {
	if fhs.ComputeMoveBalanceFeeCalled != nil {
		return fhs.ComputeMoveBalanceFeeCalled(tx)
	}
	return big.NewInt(0)
}

// ComputeTxFee -
func (fhs *FeeHandlerStub) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	if fhs.ComputeTxFeeCalled != nil {
		return fhs.ComputeTxFeeCalled(tx)
	}
	return big.NewInt(0)
}

// GenesisTotalSupply -
func (fhs *FeeHandlerStub) GenesisTotalSupply() *big.Int {
	if fhs.GenesisTotalSupplyCalled != nil {
		return fhs.GenesisTotalSupplyCalled()
	}

	return big.NewInt(0)
}

// CheckValidityTxValues -
func (fhs *FeeHandlerStub) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if fhs.CheckValidityTxValuesCalled != nil {
		return fhs.CheckValidityTxValuesCalled(tx)
	}
	return nil
}

// SplitTxGasInCategories -
func (fhs *FeeHandlerStub) SplitTxGasInCategories(tx data.TransactionWithFeeHandler) (uint64, uint64) {
	if fhs.SplitTxGasInCategoriesCalled != nil {
		return fhs.SplitTxGasInCategoriesCalled(tx)
	}
	return 0, 0
}

// GasPriceForProcessing -
func (fhs *FeeHandlerStub) GasPriceForProcessing(tx data.TransactionWithFeeHandler) uint64 {
	if fhs.GasPriceForProcessingCalled != nil {
		return fhs.GasPriceForProcessingCalled(tx)
	}
	return 0
}

// GasPriceForMove -
func (fhs *FeeHandlerStub) GasPriceForMove(tx data.TransactionWithFeeHandler) uint64 {
	if fhs.GasPriceForMoveCalled != nil {
		return fhs.GasPriceForMoveCalled(tx)
	}
	return 0
}

// MinGasPriceForProcessing -
func (fhs *FeeHandlerStub) MinGasPriceForProcessing() uint64 {
	if fhs.MinGasPriceForProcessingCalled != nil {
		return fhs.MinGasPriceForProcessingCalled()
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (fhs *FeeHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx data.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if fhs.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return fhs.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}
	return 0, big.NewInt(0)
}

// ComputeTxFeeBasedOnGasUsed -
func (fhs *FeeHandlerStub) ComputeTxFeeBasedOnGasUsed(tx data.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if fhs.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return fhs.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}
	return big.NewInt(0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhs *FeeHandlerStub) IsInterfaceNil() bool {
	return fhs == nil
}
