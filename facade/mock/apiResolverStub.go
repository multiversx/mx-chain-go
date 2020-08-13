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
	ExecuteSCQueryHandler                  func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ExecuteSCQueryWithCallerAndValueCalled func(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error)
	StatusMetricsHandler                   func() external.StatusMetricsHandler
	ComputeTransactionGasLimitHandler      func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteSCQuery -
func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryHandler(query)
}

// ExecuteQueryWithCallerAndValue -
func (ars *ApiResolverStub) ExecuteQueryWithCallerAndValue(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryWithCallerAndValueCalled(query, callerAddr, callValue)
}

// StatusMetrics -
func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	return ars.StatusMetricsHandler()
}

// ComputeTransactionGasLimit -
func (ars *ApiResolverStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	return ars.ComputeTransactionGasLimitHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
