package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ApiResolverStub -
type ApiResolverStub struct {
	ExecuteSCQueryHandler             func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler              func() external.StatusMetricsHandler
	ComputeTransactionGasLimitHandler func(tx *transaction.Transaction) (*transaction.CostResponse, error)
	GetTotalStakedValueHandler        func() (*api.StakeValues, error)
	GetDirectStakedListHandler        func() ([]*api.DirectStakedValue, error)
	GetDelegatorsListHandler          func() ([]*api.Delegator, error)
}

// ExecuteSCQuery -
func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryHandler(query)
}

// StatusMetrics -
func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	return ars.StatusMetricsHandler()
}

// ComputeTransactionGasLimit -
func (ars *ApiResolverStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return ars.ComputeTransactionGasLimitHandler(tx)
}

// GetTotalStakedValue -
func (ars *ApiResolverStub) GetTotalStakedValue() (*api.StakeValues, error) {
	return ars.GetTotalStakedValueHandler()
}

// GetDirectStakedList -
func (ars *ApiResolverStub) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	if ars.GetDirectStakedListHandler != nil {
		return ars.GetDirectStakedListHandler()
	}

	return nil, nil
}

// GetDelegatorsList -
func (ars *ApiResolverStub) GetDelegatorsList() ([]*api.Delegator, error) {
	if ars.GetDelegatorsListHandler != nil {
		return ars.GetDelegatorsListHandler()
	}

	return nil, nil
}

// Close -
func (ars *ApiResolverStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
