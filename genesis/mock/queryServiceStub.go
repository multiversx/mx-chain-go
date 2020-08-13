package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// QueryServiceStub -
type QueryServiceStub struct {
	ComputeScCallGasLimitCalled          func(tx *transaction.Transaction) (uint64, error)
	ExecuteQueryCalled                   func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ExecuteQueryWithCallerAndValueCalled func(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error)
}

// ComputeScCallGasLimit -
func (qss *QueryServiceStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	if qss.ComputeScCallGasLimitCalled != nil {
		return qss.ComputeScCallGasLimitCalled(tx)
	}

	return 0, nil
}

// ExecuteQuery -
func (qss *QueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if qss.ExecuteQueryCalled != nil {
		return qss.ExecuteQueryCalled(query)
	}

	return &vmcommon.VMOutput{}, nil
}

// ExecuteQueryWithCallerAndValue -
func (qss *QueryServiceStub) ExecuteQueryWithCallerAndValue(query *process.SCQuery, callerAddr []byte, callValue *big.Int) (*vmcommon.VMOutput, error) {
	if qss.ExecuteQueryWithCallerAndValueCalled != nil {
		return qss.ExecuteQueryWithCallerAndValueCalled(query, callerAddr, callValue)
	}

	return &vmcommon.VMOutput{}, nil
}

// IsInterfaceNil -
func (qss *QueryServiceStub) IsInterfaceNil() bool {
	return qss == nil
}
