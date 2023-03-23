package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

// ExecutionOrderHandlerStub -
type ExecutionOrderHandlerStub struct {
}

// PutExecutionOrderInTransactionPool -
func (e *ExecutionOrderHandlerStub) PutExecutionOrderInTransactionPool(
	_ *outport.TransactionPool,
	_ data.HeaderHandler,
	_ data.BodyHandler,
	_ data.HeaderHandler,
) ([]string, []string, error) {
	return nil, nil, nil
}
