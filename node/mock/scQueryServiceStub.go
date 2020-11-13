package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SCQueryServiceStub -
type SCQueryServiceStub struct {
	ExecuteQueryCalled           func(*process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimitHandler func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteQuery -
func (serviceStub *SCQueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return serviceStub.ExecuteQueryCalled(query)
}

// ComputeScCallGasLimit -
func (serviceStub *SCQueryServiceStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	return serviceStub.ComputeScCallGasLimitHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (serviceStub *SCQueryServiceStub) IsInterfaceNil() bool {
	return serviceStub == nil
}
