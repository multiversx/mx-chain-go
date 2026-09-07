package asyncExecution

import "github.com/multiversx/mx-chain-go/process/asyncExecution/cache"

// Process -
func (he *headersExecutor) Process(pair cache.HeaderBodyPair) error {
	return he.process(pair)
}
