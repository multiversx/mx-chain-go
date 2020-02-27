package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScQueryMock -
type ScQueryMock struct {
	ExecuteQueryCalled           func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeTransactionCostCalled func(tx *transaction.Transaction) (*big.Int, error)
}

// ExecuteQuery -
func (s *ScQueryMock) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryCalled != nil {
		return s.ExecuteQueryCalled(query)
	}
	return &vmcommon.VMOutput{}, nil
}

// ComputeTransactionCost --
func (s *ScQueryMock) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return s.ComputeTransactionCostCalled(tx)
}

// IsInterfaceNil -
func (s *ScQueryMock) IsInterfaceNil() bool {
	return s == nil
}
