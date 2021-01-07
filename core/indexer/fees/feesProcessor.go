package fees

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process"
)

type feesProcessor struct {
	economicsHandler process.EconomicsDataHandler
}

// NewFeesProcessor will create a new instance of fee processor
func NewFeesProcessor(economicsHandler process.EconomicsDataHandler) *feesProcessor {
	return &feesProcessor{
		economicsHandler: economicsHandler,
	}
}

// ComputeGasUsedAndFeeBasedOnRefundValue will compute gas used value and transaction fee using refund value from a SCR
func (fp *feesProcessor) ComputeGasUsedAndFeeBasedOnRefundValue(tx process.TransactionWithFeeHandler, refundValueStr string) (uint64, *big.Int) {
	refundValue, ok := big.NewInt(0).SetString(refundValueStr, 10)
	if !ok {
		refundValue = big.NewInt(0)
	}

	if refundValue.Cmp(big.NewInt(0)) == 0 {
		txFee := fp.economicsHandler.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	gasRefund := fp.computeGasRefund(refundValue, tx.GetGasPrice())
	gasUsed := tx.GetGasLimit() - gasRefund
	txFee := big.NewInt(0).Sub(fp.economicsHandler.ComputeTxFee(tx), refundValue)

	return gasUsed, txFee
}

// ComputeTxFeeBasedOnGasUsed will compute transaction fee
func (fp *feesProcessor) ComputeTxFeeBasedOnGasUsed(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	moveBalanceGasLimit := fp.economicsHandler.ComputeGasLimit(tx)
	moveBalanceFee := fp.economicsHandler.ComputeMoveBalanceFee(tx)
	if gasUsed <= moveBalanceGasLimit {
		return moveBalanceFee
	}

	computeFeeForProcessing := fp.economicsHandler.ComputeFeeForProcessing(tx, gasUsed-moveBalanceGasLimit)
	txFee := big.NewInt(0).Add(moveBalanceFee, computeFeeForProcessing)

	return txFee
}

// ComputeMoveBalanceGasUsed will compute gas used for a move balance
func (fp *feesProcessor) ComputeMoveBalanceGasUsed(tx process.TransactionWithFeeHandler) uint64 {
	return fp.economicsHandler.ComputeGasLimit(tx)
}

func (fp *feesProcessor) computeGasRefund(refundValue *big.Int, gasPrice uint64) uint64 {
	//gasPriceWithModifier := uint64(float64(gasPrice) * fp.economicsHandler.GasPriceModifier())
	//
	//gasPriceBig := big.NewInt(0).SetUint64(gasPriceWithModifier)
	//
	//gasRefund := big.NewInt(0).Div(refundValue, gasPriceBig)
	//
	//return gasRefund.Uint64()

	gasPriceBig := big.NewInt(0).SetUint64(gasPrice)

	gasRefund := big.NewInt(0).Div(refundValue, gasPriceBig)

	gasRefundFloat := big.NewFloat(0).SetInt(gasRefund)

	gasPriceModifierFloat := big.NewFloat(fp.economicsHandler.GasPriceModifier())

	gasRefundFloat.Mul(gasRefundFloat, gasPriceModifierFloat)

	gasRefundUint, _ := gasRefundFloat.Uint64()

	return gasRefundUint
}
