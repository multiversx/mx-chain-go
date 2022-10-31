package fee

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/enablers"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/external/timemachine"
	"github.com/ElrondNetwork/elrond-go/process/economics"
)

var log = logger.GetOrCreate("node/external/timemachine/fee")

type feeComputer struct {
	builtInFunctionsCostHandler economics.BuiltInFunctionsCostHandler
	economicsConfig             config.EconomicsConfig
	economicsInstances          map[uint32]economicsDataWithComputeFee
	enableEpochsHandler         common.EnableEpochsHandler
	mutex                       sync.RWMutex
}

// NewFeeComputer creates a fee computer which handles historical transactions, as well
func NewFeeComputer(args ArgsNewFeeComputer) (*feeComputer, error) {
	err := args.check()
	if err != nil {
		return nil, err
	}

	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(args.EnableEpochsConfig, &timemachine.DisabledEpochNotifier{})
	if err != nil {
		return nil, err
	}

	computer := &feeComputer{
		builtInFunctionsCostHandler: args.BuiltInFunctionsCostHandler,
		economicsConfig:             args.EconomicsConfig,
		// TODO: use a LRU cache instead
		economicsInstances:  make(map[uint32]economicsDataWithComputeFee),
		enableEpochsHandler: enableEpochsHandler,
	}

	// Create some economics data instance (but do not save them) in order to validate the arguments:
	_, err = computer.createEconomicsInstance(0)
	if err != nil {
		return nil, err
	}

	// TODO: Handle fees for guarded transactions, when enabled.

	return computer, nil
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
	epochSubscriberHandler, ok := computer.enableEpochsHandler.(core.EpochSubscriberHandler)
	if !ok {
		return nil, external.ErrEpochSubscriberHandlerWrongTypeAssertion
	}

	epochSubscriberHandler.EpochConfirmed(epoch, 0)

	args := economics.ArgsNewEconomicsData{
		Economics:                   &computer.economicsConfig,
		BuiltInFunctionsCostHandler: computer.builtInFunctionsCostHandler,
		EpochNotifier:               &timemachine.DisabledEpochNotifier{},
		EnableEpochsHandler:         computer.enableEpochsHandler,
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
