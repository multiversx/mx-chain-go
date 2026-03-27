package asyncExecution

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

// BlocksCache defines what a block queue should be able to do
type BlocksCache interface {
	GetByNonce(nonce uint64) (cache.HeaderBodyPair, bool)
	AddOrReplace(pair cache.HeaderBodyPair) error
	Remove(nonce uint64)
	GetSignalBlockAddedChan() <-chan struct{}
	IsInterfaceNil() bool
}

// ExecutionResultsHandler defines what an execution results handler should be able to do
type ExecutionResultsHandler interface {
	AddExecutionResult(executionResult data.BaseExecutionResultHandler) (bool, error)
	IsInterfaceNil() bool
}

// BlockProcessor defines what a block processor should be able to do
type BlockProcessor interface {
	ProcessBlockProposal(header data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error)
	CommitBlockProposalState(headerHandler data.HeaderHandler) error
	RevertBlockProposalState()
	PruneTrieAsyncHeader(header data.HeaderHandler)
	IsInterfaceNil() bool
}
