package fee

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/node/external/timemachine"
	"github.com/ElrondNetwork/elrond-go/process/economics"
)

var log = logger.GetOrCreate("node/external/timemachine/fee")

type feeComputer struct {
	builtInFunctionsCostHandler    economics.BuiltInFunctionsCostHandler
	economicsConfig                config.EconomicsConfig
	penalizedTooMuchGasEnableEpoch uint32
	gasPriceModifierEnableEpoch    uint32
	economicsInstances             map[int]economicsDataWithComputeFee
	mutex                          sync.RWMutex
}

// NewFeeComputer creates a fee computer which handles historical transactions, as well
func NewFeeComputer(args ArgsNewFeeComputer) (*feeComputer, error) {
	err := args.check()
	if err != nil {
		return nil, err
	}

	computer := &feeComputer{
		builtInFunctionsCostHandler:    args.BuiltInFunctionsCostHandler,
		economicsConfig:                args.EconomicsConfig,
		penalizedTooMuchGasEnableEpoch: args.PenalizedTooMuchGasEnableEpoch,
		gasPriceModifierEnableEpoch:    args.GasPriceModifierEnableEpoch,
		// TODO: use a LRU cache instead
		economicsInstances: make(map[int]economicsDataWithComputeFee),
	}

	// Create some economics data instance (but do not save them) in order to validate the arguments:
	_, err = computer.createEconomicsInstance(0)
	if err != nil {
		return nil, err
	}

	_, err = computer.createEconomicsInstance(int(args.PenalizedTooMuchGasEnableEpoch))
	if err != nil {
		return nil, err
	}

	_, err = computer.createEconomicsInstance(int(args.GasPriceModifierEnableEpoch))
	if err != nil {
		return nil, err
	}

	return computer, nil
}

// ComputeTransactionFee computes a transaction fee, at a given epoch
func (computer *feeComputer) ComputeTransactionFee(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
	instance, err := computer.getOrCreateInstance(epoch)
	if err != nil {
		log.Error("ComputeTransactionFee(): unexpected error when creating an economicsData instance", "epoch", epoch, "error", err)
		return big.NewInt(0)
	}

	fee := instance.ComputeTxFee(tx)
	return fee
}

// ComputeTransactionFeeForMoveBalance computes the "move balance" component of the transaction fee, at a given epoch
func (computer *feeComputer) ComputeTransactionFeeForMoveBalance(tx data.TransactionWithFeeHandler, epoch int) *big.Int {
	instance, err := computer.getOrCreateInstance(epoch)
	if err != nil {
		log.Error("ComputeTransactionFeeForMoveBalance(): unexpected error when creating an economicsData instance", "epoch", epoch, "error", err)
		return big.NewInt(0)
	}

	fee := instance.ComputeMoveBalanceFee(tx)
	return fee
}

// getOrCreateInstance gets or lazily creates a fee computer (using "double-checked locking" pattern)
func (computer *feeComputer) getOrCreateInstance(epoch int) (economicsDataWithComputeFee, error) {
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

func (computer *feeComputer) createEconomicsInstance(epoch int) (economicsDataWithComputeFee, error) {
	args := economics.ArgsNewEconomicsData{
		Economics:                      &computer.economicsConfig,
		PenalizedTooMuchGasEnableEpoch: computer.penalizedTooMuchGasEnableEpoch,
		GasPriceModifierEnableEpoch:    computer.gasPriceModifierEnableEpoch,
		BuiltInFunctionsCostHandler:    computer.builtInFunctionsCostHandler,
		EpochNotifier:                  &timemachine.DisabledEpochNotifier{},
	}

	economicsData, err := economics.NewEconomicsData(args)
	if err != nil {
		return nil, err
	}

	economicsData.EpochConfirmed(uint32(epoch), 0)

	return economicsData, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (computer *feeComputer) IsInterfaceNil() bool {
	return computer == nil
}
