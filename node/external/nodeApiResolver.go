package external

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
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
}

// NewNodeApiResolver creates a new nodeApiResolver instance
func NewNodeApiResolver(args ApiResolverArgs) (*nodeApiResolver, error) {
	if check.IfNil(args.ScQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(args.StatusMetrics) {
		return nil, ErrNilStatusMetrics
	}
	if check.IfNil(args.TxCostHandler) {
		return nil, ErrNilTransactionCostHandler
	}
	if check.IfNil(args.StakedValueHandler) {
		return nil, ErrNilTotalStakedValueHandler
	}
	if check.IfNil(args.VmContainer) {
		return nil, ErrNilVmContainer
	}
	if check.IfNil(args.VmFactory){
		return nil, ErrNilVmFactory
	}

	return &nodeApiResolver{
		scQueryService:          args.ScQueryService,
		statusMetricsHandler:    args.StatusMetrics,
		txCostHandler:           args.TxCostHandler,
		vmContainer:             args.VmContainer,
		vmFactory:               args.VmFactory,
		totalStakedValueHandler: args.StakedValueHandler,
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

//ComputeTransactionGasLimit will calculate how many gas a transaction will consume
func (nar *nodeApiResolver) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
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
func (nar *nodeApiResolver) GetTotalStakedValue() (*big.Int, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue()
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *nodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
