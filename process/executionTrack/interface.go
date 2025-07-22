package executionTrack

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type HeaderWithExecutionResults interface {
	data.HeaderHandler
	GetExecutionResults() []*block.ExecutionResult
}
