package asyncExecution

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

// BlocksQueue defines what a block queue should be able to do
type BlocksQueue interface {
	Pop() (queue.HeaderBodyPair, bool)
	Peek() (queue.HeaderBodyPair, bool)
	IsInterfaceNil() bool
	Close()
}

// ExecutionResultsHandler defines what an execution results handler should be able to do
type ExecutionResultsHandler interface {
	AddExecutionResult(executionResult data.ExecutionResultHandler) error
	IsInterfaceNil() bool
}

// BlockProcessor defines what a block processor should be able to do
type BlockProcessor interface {
	ProcessBlock(header data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error)
	IsInterfaceNil() bool
}
