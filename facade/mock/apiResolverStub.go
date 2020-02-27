package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ApiResolverStub -
type ApiResolverStub struct {
	ExecuteSCQueryHandler         func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler          func() external.StatusMetricsHandler
	ComputeTransactionCostHandler func(tx *transaction.Transaction) (*big.Int, error)
}

// ExecuteSCQuery -
func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryHandler(query)
}

// StatusMetrics -
func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	return ars.StatusMetricsHandler()
}

// ComputeTransactionCost -
func (ars *ApiResolverStub) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return ars.ComputeTransactionCostHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
