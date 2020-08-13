package external

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scQueryService       SCQueryService
	statusMetricsHandler StatusMetricsHandler
	txCostHandler        TransactionCostHandler
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(
	scQueryService SCQueryService,
	statusMetricsHandler StatusMetricsHandler,
	txCostHandler TransactionCostHandler,
) (*NodeApiResolver, error) {
	if check.IfNil(scQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(statusMetricsHandler) {
		return nil, ErrNilStatusMetrics
	}
	if check.IfNil(txCostHandler) {
		return nil, ErrNilTransactionCostHandler
	}

	return &NodeApiResolver{
		scQueryService:       scQueryService,
		statusMetricsHandler: statusMetricsHandler,
		txCostHandler:        txCostHandler,
	}, nil
}

// ExecuteSCQuery retrieves data stored in a SC account through a VM
func (nar *NodeApiResolver) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return nar.scQueryService.ExecuteQuery(query)
}

// ExecuteQueryWithValue retrieves data stored in a SC account through a VM but also includes the call value
func (nar *NodeApiResolver) ExecuteQueryWithValue(query *process.SCQuery, callValue *big.Int) (*vmcommon.VMOutput, error) {
	return nar.scQueryService.ExecuteQueryWithValue(query, callValue)
}

// StatusMetrics returns an implementation of the StatusMetricsHandler interface
func (nar *NodeApiResolver) StatusMetrics() StatusMetricsHandler {
	return nar.statusMetricsHandler
}

//ComputeTransactionGasLimit will calculate how many gas a transaction will consume
func (nar *NodeApiResolver) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	return nar.txCostHandler.ComputeTransactionGasLimit(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
