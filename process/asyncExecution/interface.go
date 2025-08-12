package asyncExecution

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

type BlocksQueue interface {
	Pop() (queue.HeaderBodyPair, bool)
	IsInterfaceNil() bool
	Close()
}

type ExecutionResultsHandler interface {
	AddExecutionResult(executionResult data.ExecutionResultHandler) error
	IsInterfaceNil() bool
}

type BlockProcessor interface {
	ProcessBlock(header data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error)
	IsInterfaceNil() bool
}
