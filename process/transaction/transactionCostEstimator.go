package transaction

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
)

type transactionCostEstimator struct {
	txTypeHandler      process.TxTypeHandler
	feeHandler         process.FeeHandler
	query              external.SCQueryService
	builtInCostHandler economics.BuiltInFunctionsCostHandler
	mutExecution       sync.RWMutex
	storePerByteCost   uint64
	compilePerByteCost uint64
	selfShardID        uint32
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	query external.SCQueryService,
	gasSchedule core.GasScheduleNotifier,
	builtInCostHandler economics.BuiltInFunctionsCostHandler,
	selfShardID uint32,
) (*transactionCostEstimator, error) {
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(query) {
		return nil, external.ErrNilSCQueryService
	}
	if check.IfNil(builtInCostHandler) {
		return nil, process.ErrNilBuiltInFunctionsCostHandler
	}

	compileCost, storeCost := getOperationCost(gasSchedule.LatestGasSchedule())

	return &transactionCostEstimator{
		txTypeHandler:      txTypeHandler,
		feeHandler:         feeHandler,
		query:              query,
		storePerByteCost:   compileCost,
		compilePerByteCost: storeCost,
		builtInCostHandler: builtInCostHandler,
		selfShardID:        selfShardID,
	}, nil
}

// GasScheduleChange is called when gas schedule is changed, thus all contracts must be updated
func (tce *transactionCostEstimator) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	tce.mutExecution.Lock()
	tce.compilePerByteCost, tce.storePerByteCost = getOperationCost(gasSchedule)
	tce.mutExecution.Unlock()
}

func getOperationCost(gasSchedule map[string]map[string]uint64) (uint64, uint64) {
	baseOpMap, ok := gasSchedule[core.BaseOperationCost]
	if !ok {
		return 0, 0
	}

	storeCost, ok := baseOpMap["StorePerByte"]
	if !ok {
		return 0, 0
	}

	compilerCost, ok := baseOpMap["CompilePerByte"]
	if !ok {
		return 0, 0
	}

	return storeCost, compilerCost
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	tce.mutExecution.RLock()
	defer tce.mutExecution.RUnlock()

	txType, _ := tce.txTypeHandler.ComputeTransactionType(tx)
	tx.GasPrice = 1

	switch txType {
	case process.MoveBalance:
		return &transaction.CostResponse{
			GasUnits: tce.feeHandler.ComputeGasLimit(tx),
		}, nil
	case process.SCDeployment:
		return tce.computeScDeployGasLimit(tx)
	case process.SCInvoking:
		return tce.computeScCallGasLimit(tx)
	case process.BuiltInFunctionCall:
		// TODO Built in functions with SC call not yet supported and also ESDTNFTTransfer
		return tce.computeBuiltInFunctionGasLimit(tx)
	case process.RelayedTx:
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: "cannot compute cost of the relayed transaction",
		}, nil
	default:
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: process.ErrWrongTransaction.Error(),
		}, nil
	}
}
func (tce *transactionCostEstimator) computeScDeployGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	scDeployCost := uint64(len(tx.Data)) * (tce.storePerByteCost + tce.compilePerByteCost)
	baseCost := tce.feeHandler.ComputeGasLimit(tx)

	return &transaction.CostResponse{
		GasUnits: baseCost + scDeployCost,
	}, nil
}

func (tce *transactionCostEstimator) computeBuiltInFunctionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	if tce.selfShardID == core.MetachainShardId {
		return tce.computeScCallGasLimit(tx)
	}

	gasUsed := tce.builtInCostHandler.ComputeBuiltInCost(tx)
	gasUsed += tce.feeHandler.ComputeGasLimit(tx)

	return &transaction.CostResponse{
		GasUnits: gasUsed,
	}, nil
}

func (tce *transactionCostEstimator) computeScCallGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	scCallGasLimit, err := tce.query.ComputeScCallGasLimit(tx)
	if err != nil {
		return &transaction.CostResponse{
			GasUnits:   0,
			RetMessage: err.Error(),
		}, nil
	}

	baseCost := tce.feeHandler.ComputeGasLimit(tx)
	return &transaction.CostResponse{
		GasUnits: baseCost + scCallGasLimit,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
