package fee

import (
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/external/timemachine"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/economics"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("node/external/timemachine/fee")

type feeComputer struct {
	txVersionChecker            process.TxVersionCheckerHandler
	builtInFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	economicsConfig             config.EconomicsConfig
	economicsInstances          map[uint32]economicsDataWithComputeFee
	enableEpochsConfig          config.EnableEpochs
	mutex                       sync.RWMutex
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
		// TODO: use a LRU cache instead
		economicsInstances: make(map[uint32]economicsDataWithComputeFee),
		enableEpochsConfig: args.EnableEpochsConfig,
		txVersionChecker:   args.TxVersionChecker,
	}

	// Create some economics data instance (but do not save them) in order to validate the arguments:
	_, err = computer.createEconomicsInstance(0)
	if err != nil {
		return nil, err
	}

	// TODO: Handle fees for guarded transactions, when enabled.

	return computer, nil
}

// ComputeGasUsedAndFeeBasedOnRefundValue computes gas used and fee based on the refund value, at a given epoch
func (computer *feeComputer) ComputeGasUsedAndFeeBasedOnRefundValue(tx *transaction.ApiTransactionResult, refundValue *big.Int) (uint64, *big.Int) {
	instance, err := computer.getOrCreateInstance(tx.Epoch)
	if err != nil {
		log.Error("ComputeGasUsedAndFeeBasedOnRefundValue(): unexpected error when creating an economicsData instance", "epoch", tx.Epoch, "error", err)
		return 0, big.NewInt(0)
	}

	return instance.ComputeGasUsedAndFeeBasedOnRefundValue(tx.Tx, refundValue)
}

// ComputeTxFeeBasedOnGasUsed computes fee based on gas used, at a given epoch
func (computer *feeComputer) ComputeTxFeeBasedOnGasUsed(tx *transaction.ApiTransactionResult, gasUsed uint64) *big.Int {
	instance, err := computer.getOrCreateInstance(tx.Epoch)
	if err != nil {
		log.Error("ComputeTxFeeBasedOnGasUsed(): unexpected error when creating an economicsData instance", "epoch", tx.Epoch, "error", err)
		return big.NewInt(0)
	}

	return instance.ComputeTxFeeBasedOnGasUsed(tx.Tx, gasUsed)
}

// ComputeGasLimit computes a transaction gas limit, at a given epoch
func (computer *feeComputer) ComputeGasLimit(tx *transaction.ApiTransactionResult) uint64 {
	instance, err := computer.getOrCreateInstance(tx.Epoch)
	if err != nil {
		log.Error("ComputeGasLimit(): unexpected error when creating an economicsData instance", "epoch", tx.Epoch, "error", err)
		return 0
	}

	return instance.ComputeGasLimit(tx.Tx)
}

// ComputeTransactionFee computes a transaction fee, at a given epoch
func (computer *feeComputer) ComputeTransactionFee(tx *transaction.ApiTransactionResult) *big.Int {
	instance, err := computer.getOrCreateInstance(tx.Epoch)
	if err != nil {
		log.Error("ComputeTransactionFee(): unexpected error when creating an economicsData instance", "epoch", tx.Epoch, "error", err)
		return big.NewInt(0)
	}

	return instance.ComputeTxFee(tx.Tx)
}

// getOrCreateInstance gets or lazily creates a fee computer (using "double-checked locking" pattern)
func (computer *feeComputer) getOrCreateInstance(epoch uint32) (economicsDataWithComputeFee, error) {
	computer.mutex.RLock()
	instance, ok := computer.economicsInstances[epoch]
	computer.mutex.RUnlock()
	if ok {
		return instance, nil
	}

	computer.mutex.Lock()
	defer computer.mutex.Unlock()

	instance, ok = computer.economicsInstances[epoch]
	if ok {
		return instance, nil
	}

	newInstance, err := computer.createEconomicsInstance(epoch)
	if err != nil {
		return nil, err
	}

	computer.economicsInstances[epoch] = newInstance
	return newInstance, nil
}

func (computer *feeComputer) createEconomicsInstance(epoch uint32) (economicsDataWithComputeFee, error) {
	epochNotifier := &timemachine.DisabledEpochNotifier{}
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(computer.enableEpochsConfig, epochNotifier)
	if err != nil {
		return nil, err
	}

	enableEpochsHandler.EpochConfirmed(epoch, 0)

	args := economics.ArgsNewEconomicsData{
		Economics:                   &computer.economicsConfig,
		BuiltInFunctionsCostHandler: computer.builtInFunctionsCostHandler,
		EpochNotifier:               &timemachine.DisabledEpochNotifier{},
		EnableEpochsHandler:         enableEpochsHandler,
		TxVersionChecker:            computer.txVersionChecker,
	}

	economicsData, err := economics.NewEconomicsData(args)
	if err != nil {
		return nil, err
	}

	economicsData.EpochConfirmed(epoch, 0)

	return economicsData, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (computer *feeComputer) IsInterfaceNil() bool {
	return computer == nil
}
