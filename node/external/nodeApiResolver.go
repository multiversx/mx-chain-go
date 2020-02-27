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
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(scQueryService SCQueryService, statusMetricsHandler StatusMetricsHandler) (*NodeApiResolver, error) {
	if check.IfNil(scQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(statusMetricsHandler) {
		return nil, ErrNilStatusMetrics
	}

	return &NodeApiResolver{
		scQueryService:       scQueryService,
		statusMetricsHandler: statusMetricsHandler,
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

//ComputeTransactionCost will calculate how many gas a transaction will consume
func (nar *NodeApiResolver) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return nar.scQueryService.ComputeTransactionCost(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
