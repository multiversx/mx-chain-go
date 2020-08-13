package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScQueryStub -
type ScQueryStub struct {
	ExecuteQueryCalled                   func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ExecuteQueryWithCallerAndValueCalled func(query *process.SCQuery, callerAddr []byte, value *big.Int) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimitHandler         func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteQuery -
func (s *ScQueryStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryCalled != nil {
		return s.ExecuteQueryCalled(query)
	}
	return &vmcommon.VMOutput{}, nil
}

// ExecuteQueryWithCallerAndValue -
func (s *ScQueryStub) ExecuteQueryWithCallerAndValue(query *process.SCQuery, callerAddr []byte, value *big.Int) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryWithCallerAndValueCalled != nil {
		return s.ExecuteQueryWithCallerAndValueCalled(query, callerAddr, value)
	}
	return &vmcommon.VMOutput{}, nil
}

// ComputeScCallGasLimit --
func (s *ScQueryStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	if s.ComputeScCallGasLimitHandler != nil {
		return s.ComputeScCallGasLimitHandler(tx)
	}
	return 100, nil
}

// IsInterfaceNil -
func (s *ScQueryStub) IsInterfaceNil() bool {
	return s == nil
}
