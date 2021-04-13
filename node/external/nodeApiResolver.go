package external

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ApiResolverArgs holds the arguments for a node API resolver
type ApiResolverArgs struct {
	ScQueryService     SCQueryService
	StatusMetrics      StatusMetricsHandler
	TxCostHandler      TransactionCostHandler
	VmFactory          process.VirtualMachinesContainerFactory
	VmContainer        process.VirtualMachinesContainer
	StakedValueHandler TotalStakedValueHandler
}

// nodeApiResolver can resolve API requests
type nodeApiResolver struct {
	scQueryService          SCQueryService
	statusMetricsHandler    StatusMetricsHandler
	txCostHandler           TransactionCostHandler
	vmContainer             process.VirtualMachinesContainer
	vmFactory               process.VirtualMachinesContainerFactory
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
	if check.IfNil(args.VmContainer) {
		return nil, ErrNilVmContainer
	}
	if check.IfNil(args.VmFactory) {
		return nil, ErrNilVmFactory
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
func (nar *nodeApiResolver) Close() error {
	var err1, err2 error
	if !check.IfNil(nar.vmContainer) {
		err1 = nar.vmContainer.Close()
	}
	if !check.IfNil(nar.vmFactory) {
		err2 = nar.vmFactory.Close()
	}
	if err1 != nil || err2 != nil {
		return fmt.Errorf("err closing vmContainer: %v, err closing vmFactory: %v", err1, err2)
	}
	return nil
}

// GetTotalStakedValue will return total staked value
func (nar *nodeApiResolver) GetTotalStakedValue() (*api.StakeValues, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue()
}

// GetDirectStakedList will return the list for the direct staked addresses
func (nar *NodeApiResolver) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return nar.directStakedListHandler.GetDirectStakedList()
}

// GetDelegatorsList will return the delegators list
func (nar *NodeApiResolver) GetDelegatorsList() ([]*api.Delegator, error) {
	return nar.delegatedListHandler.GetDelegatorsList()
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *nodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
