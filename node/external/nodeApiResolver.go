package external

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgNodeApiResolver represents the DTO structure used in the NewNodeApiResolver constructor
type ArgNodeApiResolver struct {
	SCQueryService          SCQueryService
	StatusMetricsHandler    StatusMetricsHandler
	TxCostHandler           TransactionCostHandler
	TotalStakedValueHandler TotalStakedValueHandler
	DirectStakedListHandler DirectStakedListHandler
	DelegatedListHandler    DelegatedListHandler
}

// nodeApiResolver can resolve API requests
type nodeApiResolver struct {
	scQueryService          SCQueryService
	statusMetricsHandler    StatusMetricsHandler
	txCostHandler           TransactionCostHandler
	totalStakedValueHandler TotalStakedValueHandler
	directStakedListHandler DirectStakedListHandler
	delegatedListHandler    DelegatedListHandler
}

// NewNodeApiResolver creates a new nodeApiResolver instance
func NewNodeApiResolver(arg ArgNodeApiResolver) (*nodeApiResolver, error) {
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

	return &nodeApiResolver{
		scQueryService:          arg.SCQueryService,
		statusMetricsHandler:    arg.StatusMetricsHandler,
		txCostHandler:           arg.TxCostHandler,
		totalStakedValueHandler: arg.TotalStakedValueHandler,
		directStakedListHandler: arg.DirectStakedListHandler,
		delegatedListHandler:    arg.DelegatedListHandler,
	}, nil
}

// ExecuteSCQuery retrieves data stored in a SC account through a VM
func (nar *nodeApiResolver) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return nar.scQueryService.ExecuteQuery(query)
}

// StatusMetrics returns an implementation of the StatusMetricsHandler interface
func (nar *nodeApiResolver) StatusMetrics() StatusMetricsHandler {
	return nar.statusMetricsHandler
}

// ComputeTransactionGasLimit will calculate how many gas a transaction will consume
func (nar *nodeApiResolver) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return nar.txCostHandler.ComputeTransactionGasLimit(tx)
}

// Close closes all underlying components
//TODO(iulian) add tests
func (nar *nodeApiResolver) Close() error {
	return nar.scQueryService.Close()
}

// GetTotalStakedValue will return total staked value
func (nar *nodeApiResolver) GetTotalStakedValue() (*api.StakeValues, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue()
}

// GetDirectStakedList will return the list for the direct staked addresses
func (nar *nodeApiResolver) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return nar.directStakedListHandler.GetDirectStakedList()
}

// GetDelegatorsList will return the delegators list
func (nar *nodeApiResolver) GetDelegatorsList() ([]*api.Delegator, error) {
	return nar.delegatedListHandler.GetDelegatorsList()
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *nodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
