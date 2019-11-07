package external

import (
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scQueryService       SCQueryService
	statusMetricsHandler StatusMetricsHandler
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(scQueryService SCQueryService, statusMetricsHandler StatusMetricsHandler) (*NodeApiResolver, error) {
	if scQueryService == nil || scQueryService.IsInterfaceNil() {
		return nil, ErrNilSCQueryService
	}
	if statusMetricsHandler == nil || statusMetricsHandler.IsInterfaceNil() {
		return nil, ErrNilStatusMetrics
	}

	return &NodeApiResolver{
		scQueryService:       scQueryService,
		statusMetricsHandler: statusMetricsHandler,
	}, nil
}

// ExecuteSCQuery retrieves data stored in a SC account through a VM
func (nar *NodeApiResolver) ExecuteSCQuery(query *smartContract.SCQuery) (*vmcommon.VMOutput, error) {
	return nar.scQueryService.ExecuteQuery(query)
}

// StatusMetrics returns an implementation of the StatusMetricsHandler interface
func (nar *NodeApiResolver) StatusMetrics() StatusMetricsHandler {
	return nar.statusMetricsHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	if nar == nil {
		return true
	}
	return false
}
