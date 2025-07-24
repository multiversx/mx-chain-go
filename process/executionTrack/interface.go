package executionTrack

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// HeaderWithExecutionResults defines what a header with execution results should be able to do
type HeaderWithExecutionResults interface {
	data.HeaderHandler
	GetExecutionResults() []*block.ExecutionResult
}
