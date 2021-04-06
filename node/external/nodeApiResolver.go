package external

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scQueryService          SCQueryService
	statusMetricsHandler    StatusMetricsHandler
	txCostHandler           TransactionCostHandler
	totalStakedValueHandler TotalStakedValueHandler
	directStakedListHandler DirectStakedListHandler
	delegatedListHandler    DelegatedListHandler
}

// ArgNodeApiResolver represents the DTO structure used in the NewNodeApiResolver constructor
type ArgNodeApiResolver struct {
	SCQueryService          SCQueryService
	StatusMetricsHandler    StatusMetricsHandler
	TxCostHandler           TransactionCostHandler
	TotalStakedValueHandler TotalStakedValueHandler
	DirectStakedListHandler DirectStakedListHandler
	DelegatedListHandler    DelegatedListHandler
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(arg ArgNodeApiResolver) (*NodeApiResolver, error) {
	if check.IfNil(arg.SCQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(arg.StatusMetricsHandler) {
		return nil, ErrNilStatusMetrics
	}
	if check.IfNil(arg.TxCostHandler) {
		return nil, ErrNilTransactionCostHandler
	}
	if check.IfNil(arg.TotalStakedValueHandler) {
		return nil, ErrNilTotalStakedValueHandler
	}
	if check.IfNil(arg.DirectStakedListHandler) {
		return nil, ErrNilDirectStakeListHandler
	}
	if check.IfNil(arg.DelegatedListHandler) {
		return nil, ErrNilDelegatedListHandler
	}

	return &NodeApiResolver{
		scQueryService:          arg.SCQueryService,
		statusMetricsHandler:    arg.StatusMetricsHandler,
		txCostHandler:           arg.TxCostHandler,
		totalStakedValueHandler: arg.TotalStakedValueHandler,
		directStakedListHandler: arg.DirectStakedListHandler,
		delegatedListHandler:    arg.DelegatedListHandler,
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

// ComputeTransactionGasLimit will calculate how many gas a transaction will consume
func (nar *NodeApiResolver) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return nar.txCostHandler.ComputeTransactionGasLimit(tx)
}

// GetTotalStakedValue will return total staked value
func (nar *NodeApiResolver) GetTotalStakedValue() (*api.StakeValues, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue()
}

// GetDirectStakedList will output the list for the direct staked addresses
func (nar *NodeApiResolver) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return nar.directStakedListHandler.GetDirectStakedList()
}

// GetDelegatorsList will output the delegators list
func (nar *NodeApiResolver) GetDelegatorsList() ([]*api.Delegator, error) {
	return nar.delegatedListHandler.GetDelegatorsList()
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
