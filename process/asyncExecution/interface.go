package asyncExecution

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

// BlocksQueue defines what a block queue should be able to do
type BlocksQueue interface {
	Pop() (pair queue.HeaderBodyPair, poppedValidPair bool, shouldContinue bool)
	Peek() (queue.HeaderBodyPair, bool)
	IsInterfaceNil() bool
	Close()
}

// ExecutionResultsHandler defines what an execution results handler should be able to do
type ExecutionResultsHandler interface {
	AddExecutionResult(executionResult data.BaseExecutionResultHandler) error
	IsInterfaceNil() bool
}

// BlockProcessor defines what a block processor should be able to do
type BlockProcessor interface {
	ProcessBlockProposal(header data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error)
	IsInterfaceNil() bool
}
