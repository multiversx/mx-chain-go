package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// SCQueryServiceStub -
type SCQueryServiceStub struct {
	ExecuteQueryCalled            func(*process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeTransactionCostHandler func(tx *transaction.Transaction) (*big.Int, error)
}

// ExecuteQuery -
func (serviceStub *SCQueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return serviceStub.ExecuteQueryCalled(query)
}

// ComputeTransactionCost -
func (serviceStub *SCQueryServiceStub) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return serviceStub.ComputeTransactionCostHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (serviceStub *SCQueryServiceStub) IsInterfaceNil() bool {
	return serviceStub == nil
}
