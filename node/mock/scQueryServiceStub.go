package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SCQueryServiceStub -
type SCQueryServiceStub struct {
	ExecuteQueryCalled                   func(*process.SCQuery) (*vmcommon.VMOutput, error)
	ExecuteQueryWithCallerAndValueCalled func(query *process.SCQuery, callerAddr []byte, value *big.Int) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimitHandler         func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteQuery -
func (serviceStub *SCQueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return serviceStub.ExecuteQueryCalled(query)
}

// ExecuteQueryWithCallerAndValue -
func (serviceStub *SCQueryServiceStub) ExecuteQueryWithCallerAndValue(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error) {
	return serviceStub.ExecuteQueryWithCallerAndValueCalled(query, callerAddr, callValue)
}

// ComputeScCallGasLimit -
func (serviceStub *SCQueryServiceStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	return serviceStub.ComputeScCallGasLimitHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (serviceStub *SCQueryServiceStub) IsInterfaceNil() bool {
	return serviceStub == nil
}
