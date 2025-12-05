package asyncExecution

import "github.com/multiversx/mx-chain-go/process/asyncExecution/queue"

// Process -
func (he *headersExecutor) Process(pair queue.HeaderBodyPair) error {
	return he.process(pair)
}
