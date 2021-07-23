package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScQueryStub -
type ScQueryStub struct {
	ExecuteQueryCalled          func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimitCalled func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteQuery -
func (s *ScQueryStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryCalled != nil {
		return s.ExecuteQueryCalled(query)
	}
	return &vmcommon.VMOutput{}, nil
}

// ComputeScCallGasLimit --
func (s *ScQueryStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	if s.ComputeScCallGasLimitCalled != nil {
		return s.ComputeScCallGasLimitCalled(tx)
	}
	return 100, nil
}

// IsInterfaceNil -
func (s *ScQueryStub) IsInterfaceNil() bool {
	return s == nil
}
