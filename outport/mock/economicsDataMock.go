package mock

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("mock/economicsData")

const (
	minGasLimit      = uint64(50000)
	gasPerDataByte   = uint64(1500)
	gasPriceModifier = float64(0.01)
)

// EconomicsHandlerMock -
type EconomicsHandlerMock struct {
}

// MaxGasLimitPerBlock -
func (e *EconomicsHandlerMock) MaxGasLimitPerBlock(_ uint32) uint64 {
	return 0
}

// MinGasLimit -
func (e *EconomicsHandlerMock) MinGasLimit() uint64 {
	return minGasLimit
}

// ComputeGasLimit -
func (e *EconomicsHandlerMock) ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64 {
	gasLimit := minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * gasPerDataByte

	return gasLimit
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (e *EconomicsHandlerMock) ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if refundValue.Cmp(big.NewInt(0)) == 0 {
		txFee := e.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	txFee := e.ComputeTxFee(tx)

	txFee = big.NewInt(0).Sub(txFee, refundValue)

	moveBalanceGasUnits := e.ComputeGasLimit(tx)
	moveBalanceFee := e.computeMoveBalanceFee(tx)

	scOpFee := big.NewInt(0).Sub(txFee, moveBalanceFee)
	gasPriceForProcessing := big.NewInt(0).SetUint64(e.gasPriceForProcessing(tx))
	scOpGasUnits := big.NewInt(0).Div(scOpFee, gasPriceForProcessing)

	gasUsed := moveBalanceGasUnits + scOpGasUnits.Uint64()

	return gasUsed, txFee
}

// ComputeTxFeeBasedOnGasUsed -
func (e *EconomicsHandlerMock) ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	moveBalanceGasLimit := e.ComputeGasLimit(tx)
	moveBalanceFee := e.computeMoveBalanceFee(tx)
	if gasUsed <= moveBalanceGasLimit {
		return moveBalanceFee
	}

	computeFeeForProcessing := e.computeFeeForProcessing(tx, gasUsed-moveBalanceGasLimit)
	txFee := big.NewInt(0).Add(moveBalanceFee, computeFeeForProcessing)

	return txFee
}

func (e *EconomicsHandlerMock) computeMoveBalanceFee(tx coreData.TransactionWithFeeHandler) *big.Int {
	return core.SafeMul(tx.GetGasPrice(), e.ComputeGasLimit(tx))
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *EconomicsHandlerMock) IsInterfaceNil() bool {
	return e == nil
}

func (e *EconomicsHandlerMock) computeFeeForProcessing(tx coreData.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	gasPrice := e.gasPriceForProcessing(tx)
	return core.SafeMul(gasPrice, gasToUse)
}

func (e *EconomicsHandlerMock) gasPriceForProcessing(tx coreData.TransactionWithFeeHandler) uint64 {
	return uint64(float64(tx.GetGasPrice()) * gasPriceModifier)
}

func (e *EconomicsHandlerMock) ComputeTxFee(tx coreData.TransactionWithFeeHandler) *big.Int {
	gasLimitForMoveBalance, difference := e.splitTxGasInCategories(tx)
	moveBalanceFee := core.SafeMul(tx.GetGasPrice(), gasLimitForMoveBalance)
	if tx.GetGasLimit() <= gasLimitForMoveBalance {
		return moveBalanceFee
	}

	extraFee := e.computeFeeForProcessing(tx, difference)
	moveBalanceFee.Add(moveBalanceFee, extraFee)
	return moveBalanceFee
}

func (e *EconomicsHandlerMock) splitTxGasInCategories(tx coreData.TransactionWithFeeHandler) (gasLimitMove, gasLimitProcess uint64) {
	var err error
	gasLimitMove = e.ComputeGasLimit(tx)
	gasLimitProcess, err = core.SafeSubUint64(tx.GetGasLimit(), gasLimitMove)
	if err != nil {
		log.Error("SplitTxGasInCategories - insufficient gas for move",
			"providedGas", tx.GetGasLimit(),
			"computedMinimumRequired", gasLimitMove,
			"dataLen", len(tx.GetData()),
		)
	}

	return
}
