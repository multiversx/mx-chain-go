package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ApiResolverStub -
type ApiResolverStub struct {
	ExecuteSCQueryHandler             func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler              func() external.StatusMetricsHandler
	ComputeTransactionGasLimitHandler func(tx *transaction.Transaction) (uint64, error)
	GetTotalStakedValueHandler        func() (*big.Int, error)
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
func (ars *ApiResolverStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	return ars.ComputeTransactionGasLimitHandler(tx)
}

// GetTotalStakedValue -
func (ars *ApiResolverStub) GetTotalStakedValue() (*big.Int, error) {
	return ars.GetTotalStakedValueHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
