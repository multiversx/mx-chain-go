package external

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ApiResolverArgs holds the arguments for a node API resolver
type ApiResolverArgs struct {
	ScQueryService SCQueryService
	StatusMetrics  StatusMetricsHandler
	TxCostHandler  TransactionCostHandler
	VmFactory      process.VirtualMachinesContainerFactory
	VmContainer    process.VirtualMachinesContainer
}

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scQueryService       SCQueryService
	statusMetricsHandler StatusMetricsHandler
	txCostHandler        TransactionCostHandler
	vmContainer          process.VirtualMachinesContainer
	vmFactory            process.VirtualMachinesContainerFactory
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(args ApiResolverArgs) (*NodeApiResolver, error) {
	if check.IfNil(args.ScQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(args.StatusMetrics) {
		return nil, ErrNilStatusMetrics
	}
	if check.IfNil(args.TxCostHandler) {
		return nil, ErrNilTransactionCostHandler
	}

	return &NodeApiResolver{
		scQueryService:       args.ScQueryService,
		statusMetricsHandler: args.StatusMetrics,
		txCostHandler:        args.TxCostHandler,
		vmContainer:          args.VmContainer,
		vmFactory:            args.VmFactory,
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

// Close closes all underlying components
func (nar *NodeApiResolver) Close() error {
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

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
