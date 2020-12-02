package external

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
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
