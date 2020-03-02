package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScQueryStub -
type ScQueryStub struct {
	ExecuteQueryCalled           func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeTransactionCostCalled func(tx *transaction.Transaction) (*big.Int, error)
}

// ExecuteQuery -
func (s *ScQueryStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryCalled != nil {
		return s.ExecuteQueryCalled(query)
	}
	return &vmcommon.VMOutput{}, nil
}

// ComputeTransactionCost --
func (s *ScQueryStub) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return s.ComputeTransactionCostCalled(tx)
}

// IsInterfaceNil -
func (s *ScQueryStub) IsInterfaceNil() bool {
	return s == nil
}
