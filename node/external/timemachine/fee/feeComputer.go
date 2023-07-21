package fee

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/external/timemachine"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
)

type feeComputer struct {
	txVersionChecker            process.TxVersionCheckerHandler
	builtInFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	economicsConfig             config.EconomicsConfig
	economicsInstance           economicsDataWithComputeFee
	enableEpochsHandler         common.EnableEpochsHandler
}

// NewFeeComputer creates a fee computer which handles historical transactions, as well
func NewFeeComputer(args ArgsNewFeeComputer) (*feeComputer, error) {
	err := args.check()
	if err != nil {
		return nil, err
	}

	computer := &feeComputer{
		builtInFunctionsCostHandler: args.BuiltInFunctionsCostHandler,
		economicsConfig:             args.EconomicsConfig,
		enableEpochsHandler:         args.EnableEpochsHandler,
		txVersionChecker:            args.TxVersionChecker,
	}

	computer.economicsInstance, err = computer.createEconomicsInstance()
	if err != nil {
		return nil, err
	}

	// TODO: Handle fees for guarded transactions, when enabled.

	return computer, nil
}

// ComputeGasUsedAndFeeBasedOnRefundValue computes gas used and fee based on the refund value, at a given epoch
func (computer *feeComputer) ComputeGasUsedAndFeeBasedOnRefundValue(tx *transaction.ApiTransactionResult, refundValue *big.Int) (uint64, *big.Int) {
	return computer.economicsInstance.ComputeGasUsedAndFeeBasedOnRefundValueInEpoch(tx.Tx, refundValue, tx.Epoch)
}

// ComputeTxFeeBasedOnGasUsed computes fee based on gas used, at a given epoch
func (computer *feeComputer) ComputeTxFeeBasedOnGasUsed(tx *transaction.ApiTransactionResult, gasUsed uint64) *big.Int {
	return computer.economicsInstance.ComputeTxFeeBasedOnGasUsedInEpoch(tx.Tx, gasUsed, tx.Epoch)
}

// ComputeGasLimit computes a transaction gas limit, at a given epoch
func (computer *feeComputer) ComputeGasLimit(tx *transaction.ApiTransactionResult) uint64 {
	return computer.economicsInstance.ComputeGasLimit(tx.Tx)
}

// ComputeTransactionFee computes a transaction fee, at a given epoch
func (computer *feeComputer) ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int {
	return computer.economicsInstance.ComputeTxFeeInEpoch(tx.Tx, tx.Epoch)
}

func (computer *feeComputer) createEconomicsInstance() (economicsDataWithComputeFee, error) {
	args := economics.ArgsNewEconomicsData{
		Economics:                   &computer.economicsConfig,
		BuiltInFunctionsCostHandler: computer.builtInFunctionsCostHandler,
		EpochNotifier:               &timemachine.DisabledEpochNotifier{},
		EnableEpochsHandler:         computer.enableEpochsHandler,
		TxVersionChecker:            computer.txVersionChecker,
	}

	economicsData, err := economics.NewEconomicsData(args)
	if err != nil {
		return nil, err
	}

	economicsData.EpochConfirmed(0, 0)

	return economicsData, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (computer *feeComputer) IsInterfaceNil() bool {
	return computer == nil
}
