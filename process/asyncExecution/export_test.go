package asyncExecution

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

// Process -
func (he *headersExecutor) Process(pair queue.HeaderBodyPair) error {
	return he.process(pair)
}

// SetLastNotarizedResultIfNeeded -
func (he *headersExecutor) SetLastNotarizedResultIfNeeded(
	header data.HeaderHandler,
) error {
	return he.setLastNotarizedResultIfNeeded(header)
}
