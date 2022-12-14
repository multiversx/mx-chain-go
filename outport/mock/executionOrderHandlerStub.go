package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
)

// ExecutionOrderHandlerStub -
type ExecutionOrderHandlerStub struct {
}

// PutExecutionOrderInTransactionPool -
func (e *ExecutionOrderHandlerStub) PutExecutionOrderInTransactionPool(
	_ *outport.Pool,
	_ data.HeaderHandler,
	_ data.BodyHandler,
	_ data.HeaderHandler,
) error {
	return nil
}
